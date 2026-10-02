package services

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultAIQueryRowLimit is injected as a LIMIT clause when the generated
// query doesn't specify one, and is also used to hard-truncate results
// in Go regardless of what the query returned. See ValidateAIGeneratedSQL
// and TruncateResults.
const DefaultAIQueryRowLimit = 500

// deniedKeywords is a defense-in-depth deny-list of statement types and
// functions that must never appear in AI-generated SQL. This is NOT the
// primary safety boundary — a regex over SQL is bypassable by a
// sufficiently adversarial query. The real backstop is that these queries
// execute under a read-only Postgres role (see db/readonly_role.sql) that
// physically cannot mutate data or call these functions regardless of what
// slips past this check.
var deniedKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "GRANT",
	"REVOKE", "CREATE", "COPY", "CALL", "EXECUTE", "VACUUM", "REINDEX",
	"PG_SLEEP", "PG_READ_FILE", "PG_READ_BINARY_FILE", "PG_TERMINATE_BACKEND",
	"PG_CANCEL_BACKEND", "LO_IMPORT", "LO_EXPORT", "DBLINK",
}

var wordBoundaryCache = map[string]*regexp.Regexp{}

func wordBoundaryPattern(word string) *regexp.Regexp {
	if re, ok := wordBoundaryCache[word]; ok {
		return re
	}
	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
	wordBoundaryCache[word] = re
	return re
}

// GuardrailError is returned when generated SQL fails validation. The
// message is safe to surface to the end user (via the API's ErrorResponse
// shape); Detail is for server-side logs only.
type GuardrailError struct {
	Message string
	Detail  string
}

func (e *GuardrailError) Error() string {
	return e.Message
}

// ValidateAIGeneratedSQL enforces the app-level guardrail layer described
// in the AI chatbot plan: SELECT/WITH only, single statement, no denied
// keywords, and no references to tables outside AIQueryableTables. On
// success it returns the (possibly LIMIT-augmented) query to execute.
//
// This validation is defense-in-depth, not the primary safety mechanism —
// callers MUST execute the returned query under a read-only DB role.
func ValidateAIGeneratedSQL(sql string) (string, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return "", &GuardrailError{Message: "The generated query was empty.", Detail: "empty SQL"}
	}

	// Collapse a single trailing semicolon so we can safely detect a second
	// (multi-statement) one below.
	noTrailingSemicolon := strings.TrimSuffix(strings.TrimRight(trimmed, " \t\n"), ";")
	if strings.Contains(noTrailingSemicolon, ";") {
		return "", &GuardrailError{
			Message: "I couldn't safely answer that question.",
			Detail:  fmt.Sprintf("multi-statement query rejected: %s", sql),
		}
	}

	upperLeading := strings.ToUpper(strings.TrimSpace(noTrailingSemicolon))
	if !strings.HasPrefix(upperLeading, "SELECT") && !strings.HasPrefix(upperLeading, "WITH") {
		return "", &GuardrailError{
			Message: "I couldn't safely answer that question.",
			Detail:  fmt.Sprintf("non-SELECT query rejected: %s", sql),
		}
	}

	for _, kw := range deniedKeywords {
		if wordBoundaryPattern(kw).MatchString(noTrailingSemicolon) {
			return "", &GuardrailError{
				Message: "I couldn't safely answer that question.",
				Detail:  fmt.Sprintf("denied keyword %q found in query: %s", kw, sql),
			}
		}
	}

	if err := validateTableAllowList(noTrailingSemicolon); err != nil {
		return "", err
	}

	if err := validateAgentsColumnAccess(noTrailingSemicolon); err != nil {
		return "", err
	}

	return injectRowLimit(noTrailingSemicolon), nil
}

// ConfidentialDataMessage is the user-facing decline message for any
// attempt (natural-language pre-check in ai.handler.go, or this file's
// validateAgentsColumnAccess) to extract credentials/secrets through Bloom.
// Shared as a constant so the handler can tag log lines as security-relevant
// by comparing against it, rather than string-matching the message.
const ConfidentialDataMessage = "I can't share confidential information like passwords, tokens, or other account secrets."

// sensitiveAgentColumns lists agents columns that must never be selected by
// AI-generated queries, even though the agents table itself is allow-listed
// for legitimate lookups (e.g. resolving leads.sales1_person_id to a name).
// password/password_hashed are literal credentials; details is a jsonb blob
// that embeds the TOTP secret alongside branch/phone (see
// models.AgentDetails) — AISchemaContext deliberately never mentions any of
// these three, but the model isn't a trusted boundary, so this enforces it.
var sensitiveAgentColumns = []string{"password", "password_hashed", "details"}

// agentsTablePattern detects any reference to the agents table (FROM/JOIN,
// optionally quoted) so the check below only applies when a query actually
// touches it.
var agentsTablePattern = regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+"?agents"?\b`)

// wildcardSelectPattern matches "SELECT *" or "SELECT <alias>.*" — either
// would pull every column, sensitive ones included, off a query that also
// references agents.
var wildcardSelectPattern = regexp.MustCompile(`(?i)SELECT\s+(?:[a-zA-Z_][a-zA-Z0-9_]*\.)?\*`)

// validateAgentsColumnAccess is a narrower allow-list layered on top of
// validateTableAllowList: agents is a legitimate join target, but the
// table-level check alone would let a query select password/password_hashed
// or the TOTP-secret-bearing details column straight out of the read-only
// role. Same tradeoff as the rest of this file — coarse text matching as a
// secondary guard, not a SQL parser. The real backstop is that the
// read-only role's grant on agents is column-scoped and excludes exactly
// these columns (see readonly_role.sql).
func validateAgentsColumnAccess(sql string) error {
	if !agentsTablePattern.MatchString(sql) {
		return nil
	}
	if wildcardSelectPattern.MatchString(sql) {
		return &GuardrailError{
			Message: ConfidentialDataMessage,
			Detail:  fmt.Sprintf("wildcard select against agents rejected: %s", sql),
		}
	}
	for _, col := range sensitiveAgentColumns {
		if wordBoundaryPattern(col).MatchString(sql) {
			return &GuardrailError{
				Message: ConfidentialDataMessage,
				Detail:  fmt.Sprintf("query references restricted agents column %q: %s", col, sql),
			}
		}
	}
	return nil
}

// nonTableFromFuncPattern matches SQL built-in functions that use the
// keyword FROM for something other than a table reference — e.g.
// EXTRACT(MONTH FROM created_at), SUBSTRING(x FROM 1 FOR 3),
// TRIM(BOTH ' ' FROM x), OVERLAY(x PLACING 'y' FROM 1). Without stripping
// these first, validateTableAllowList's FROM-scanner misreads
// "created_at" (the second FROM's argument) as an attempted table
// reference and rejects an entirely safe query — this bit a real question
// ("how many leads were created in June") during manual testing.
var nonTableFromFuncPattern = regexp.MustCompile(`(?i)\b(?:EXTRACT|SUBSTRING|TRIM|OVERLAY)\s*\([^()]*\)`)

// validateTableAllowList is a coarse check: every identifier that follows
// FROM or JOIN must be one of AIQueryableTables or a CTE name defined by
// this same query's WITH clause. This intentionally tolerates false
// negatives (e.g. it won't catch every possible SQL obfuscation) since
// it's a secondary guard behind the read-only role, not a SQL parser.
func validateTableAllowList(sql string) error {
	fromJoinPattern := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+"?([a-zA-Z_][a-zA-Z0-9_]*)"?`)
	// Only used to locate table-reference FROM/JOIN keywords — the
	// original, unmodified sql is still what gets executed.
	scanTarget := nonTableFromFuncPattern.ReplaceAllString(sql, "")
	matches := fromJoinPattern.FindAllStringSubmatch(scanTarget, -1)

	allowed := make(map[string]bool, len(AIQueryableTables))
	for _, t := range AIQueryableTables {
		allowed[strings.ToLower(t)] = true
	}
	for _, cte := range cteNames(sql) {
		allowed[cte] = true
	}

	for _, m := range matches {
		table := strings.ToLower(m[1])
		if !allowed[table] {
			return &GuardrailError{
				Message: "I couldn't safely answer that question.",
				Detail:  fmt.Sprintf("query references disallowed table %q: %s", table, sql),
			}
		}
	}
	return nil
}

// cteNamePattern matches "<name> AS (" — how every CTE in a WITH clause is
// introduced, whether or not the query uses WITH RECURSIVE or defines
// multiple comma-separated CTEs.
var cteNamePattern = regexp.MustCompile(`(?i)\b([a-zA-Z_][a-zA-Z0-9_]*)\s+AS\s*\(`)

func cteNames(sql string) []string {
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "WITH") {
		return nil
	}
	matches := cteNamePattern.FindAllStringSubmatch(sql, -1)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, strings.ToLower(m[1]))
	}
	return names
}

var limitPattern = regexp.MustCompile(`(?i)\bLIMIT\s+\d+`)

// injectRowLimit appends a LIMIT clause if the query doesn't already have
// one. If it does, the existing value is left as-is — TruncateResults
// enforces the hard cap in Go regardless.
func injectRowLimit(sql string) string {
	if limitPattern.MatchString(sql) {
		return sql
	}
	return fmt.Sprintf("%s LIMIT %d", sql, DefaultAIQueryRowLimit)
}

// TruncateResults is the Go-side backstop for the row limit: even if a
// query's own LIMIT clause (or lack thereof) somehow yields more rows than
// expected, the response sent to the client and to the LLM phrasing step
// never exceeds DefaultAIQueryRowLimit rows.
func TruncateResults(results []map[string]interface{}) []map[string]interface{} {
	if len(results) <= DefaultAIQueryRowLimit {
		return results
	}
	return results[:DefaultAIQueryRowLimit]
}
