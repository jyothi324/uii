package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/db"
	"github.com/tuliporg/agent-management-api/middleware"
	"github.com/tuliporg/agent-management-api/services"
)

// aiOverallTimeout bounds the whole request: SQL generation, the DB query,
// and (sometimes) a second Gemini call to phrase the answer. Gemini
// generation is the slow, variable step — gemini-3.5-flash does internal
// reasoning before answering, which can occasionally take several
// seconds — so this needs real headroom, not the tight budget the DB
// query itself gets.
const aiOverallTimeout = 25 * time.Second

// aiDBQueryTimeout bounds only the database round trip (both the
// Postgres-side statement_timeout and a Go-side sub-context — see
// runQuery). Queries against these tables should be fast; this stays
// tight independent of how generous aiOverallTimeout is.
const aiDBQueryTimeout = 5 * time.Second

// AIHandler answers natural-language questions about a tenant's CRM data
// via text-to-SQL: Gemini generates a read-only SELECT from the question,
// the guardrail in services/sql_guardrail.go validates it, it's executed
// against the tenant DB, and the result is either formatted directly or
// (for multi-row/complex results) phrased into a sentence by a second
// Gemini call.
type AIHandler struct {
	gemini services.GeminiClient
}

func NewAIHandler(gemini services.GeminiClient) *AIHandler {
	return &AIHandler{gemini: gemini}
}

type askRequest struct {
	Question string `json:"question" binding:"required"`
}

type askResponse struct {
	Answer   string `json:"answer"`
	SQL      string `json:"sql"`
	RowCount int    `json:"row_count"`
}

// Ask handles POST /api/ai/ask. Registered after middleware.AuthMiddleware
// (routes/routes.go), so it's JWT-gated. Unlike other handlers, this one
// does NOT use middleware.GetDB(c) (the app's normal read-write tenant
// connection) — generated queries execute through db.GetAITenantManager()
// instead, which connects using the dedicated read-only role (see
// db/ai_tenant_manager.go and db/readonly_role.sql). middleware.GetTenant(c)
// gives the same tenant identifier DatabaseMiddleware already resolved.
//
// The confidential-data check runs before SQL generation, not after: the
// SQL-level guardrail (validateAgentsColumnAccess) is a real backstop, but
// AISchemaContext never mentions credential columns, so Gemini usually
// dodges rather than queries them for a question like "what is agent X's
// password" — producing a misleading "no such data" answer instead of an
// honest refusal. Catching intent in the raw question fixes that.
func (h *AIHandler) Ask(c *gin.Context) {
	var req askRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "A question is required.",
			"code":    "VALIDATION_ERROR",
			"details": err.Error(),
		})
		return
	}

	if services.IsConfidentialDataRequest(req.Question) {
		logSecurityEvent(c, "confidential-data question blocked before generation", req.Question)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": services.ConfidentialDataMessage,
			"code":  "AI_CONFIDENTIAL_DATA_REQUESTED",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), aiOverallTimeout)
	defer cancel()

	generation, err := h.gemini.GenerateSQL(ctx, req.Question)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "I couldn't come up with an answer to that just now.",
			"code":    "AI_GENERATION_FAILED",
			"details": err.Error(),
		})
		return
	}

	validatedSQL, err := services.ValidateAIGeneratedSQL(generation.SQL)
	if err != nil {
		var guardrailErr *services.GuardrailError
		status := http.StatusUnprocessableEntity
		message := "I couldn't safely answer that question."
		detail := err.Error()
		if errors.As(err, &guardrailErr) {
			message = guardrailErr.Message
			detail = guardrailErr.Detail
		}
		// detail (the rejected SQL, including real column/table names) is
		// deliberately not sent to the client — GuardrailError.Detail is
		// documented as server-side-logs-only. logSecurityEvent below is
		// what makes that documented contract actually true; previously
		// this branch logged nothing at all and sent detail to the client
		// instead, the opposite of the intent.
		logMessage := detail
		if message == services.ConfidentialDataMessage {
			logSecurityEvent(c, "confidential agents column blocked in generated SQL", logMessage)
		} else {
			log.Printf("[Bloom] guardrail rejected generated SQL — tenant=%s: %s", middleware.GetTenant(c), logMessage)
		}
		c.JSON(status, gin.H{
			"error": message,
			"code":  "AI_GUARDRAIL_REJECTED",
		})
		return
	}

	results, err := h.runQuery(ctx, c, validatedSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "I couldn't fetch that data just now.",
			"code":    "DATABASE_ERROR",
			"details": err.Error(),
		})
		return
	}
	results = services.TruncateResults(results)

	answer, err := h.formatAnswer(ctx, req.Question, results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "I found the data but couldn't phrase an answer.",
			"code":    "AI_PHRASING_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, askResponse{
		Answer:   answer,
		SQL:      validatedSQL,
		RowCount: len(results),
	})
}

// logSecurityEvent flags an attempt to extract confidential data through
// Bloom — distinct from ordinary guardrail noise (a disallowed table, a
// malformed query) — with a "SECURITY" tag so it's easy to alert/grep on,
// and the tenant + requesting agent attached so a repeated attempt can be
// traced back to who made it. reason is free text (e.g. "SELECT password
// FROM agents"); it must never be shown to the end user, only logged.
func logSecurityEvent(c *gin.Context, summary, reason string) {
	userID, userName := middleware.GetRequestUser(c)
	log.Printf("[SECURITY][Bloom] %s — tenant=%s user=%s (%s): %s",
		summary, middleware.GetTenant(c), userName, userID, reason)
}

// runQuery executes the validated, read-only query in its own transaction
// (rolled back unconditionally — it's SELECT-only, and SET LOCAL only
// scopes to the transaction anyway) with both a Postgres-side
// statement_timeout and its own tighter Go-side sub-context in effect —
// deliberately not the full aiOverallTimeout budget, since a stuck/slow
// query shouldn't be able to consume the whole request's time at the
// DB's expense of the LLM steps. The connection comes from the
// AI-specific read-only tenant manager, not the app's normal
// read-write pool.
func (h *AIHandler) runQuery(ctx context.Context, c *gin.Context, validatedSQL string) ([]map[string]interface{}, error) {
	manager := db.GetAITenantManager()
	if manager == nil {
		return nil, fmt.Errorf("AI database manager not initialized")
	}

	tenant := middleware.GetTenant(c)
	tenantDB, err := manager.GetTenantDB(tenant)
	if err != nil {
		return nil, err
	}

	// A child of ctx: effectively bounded by whichever is sooner, this
	// timeout or whatever's left of the overall request budget.
	dbCtx, cancel := context.WithTimeout(ctx, aiDBQueryTimeout)
	defer cancel()

	tx := tenantDB.GetUnderlying().WithContext(dbCtx).Begin()
	defer tx.Rollback()

	if err := tx.Exec(fmt.Sprintf("SET LOCAL statement_timeout = '%d'", aiDBQueryTimeout.Milliseconds())).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	if err := tx.Raw(validatedSQL).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// formatAnswer implements the two-step flow's conditional second call: a
// single scalar/aggregate row is formatted in Go for free; anything else
// (multi-row, multi-column, or empty) is phrased by Gemini.
func (h *AIHandler) formatAnswer(ctx context.Context, question string, results []map[string]interface{}) (string, error) {
	if scalar, ok := formatScalarAggregate(results); ok {
		return scalar, nil
	}

	// A zero-row result is exactly the case where a phrasing LLM is most
	// likely to fabricate a plausible-sounding answer instead of admitting
	// it has nothing to work with (confirmed in practice — see the "which
	// profile has maximum number of leads" incident). Never hand Gemini an
	// empty result set to "phrase" — there is nothing legitimate for it to
	// say beyond this fixed message.
	if len(results) == 0 {
		return "I couldn't find any data matching that question.", nil
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return h.gemini.PhraseAnswer(ctx, question, string(resultsJSON))
}

// formatScalarAggregate handles the common "how many"/"how much" case: a
// single row with a single numeric column (e.g. a COUNT/SUM result). This
// avoids a second Gemini call entirely for the majority of expected
// questions, at the cost of a somewhat generic sentence — good enough for
// v1 given how much of the target question set ("how many leads...",
// "how many jobs...") reduces to exactly this shape.
func formatScalarAggregate(results []map[string]interface{}) (string, bool) {
	if len(results) != 1 {
		return "", false
	}
	row := results[0]
	if len(row) != 1 {
		return "", false
	}
	for col, v := range row {
		switch strings.ToLower(col) {
		case "count", "total", "sum", "amount":
			return fmt.Sprintf("The answer is %v.", v), true
		default:
			return fmt.Sprintf("%s: %v.", col, v), true
		}
	}
	return "", false
}
