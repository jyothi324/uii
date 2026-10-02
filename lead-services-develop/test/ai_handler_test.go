package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tuliporg/agent-management-api/handlers"
	"github.com/tuliporg/agent-management-api/services"
)

// mockGeminiClient implements services.GeminiClient with canned responses,
// so these tests exercise the handler's validation/guardrail/error-handling
// logic without making real network calls or requiring a live Postgres.
//
// Note: a full happy-path test (guardrail passes, query executes, answer is
// returned) additionally requires a live tenant database, since Ask() routes
// query execution through db.GetAITenantManager() rather than an injectable
// mock DB — that path is covered by the plan's manual curl verification
// steps against a running instance, not here.
type mockGeminiClient struct {
	sql         string
	explanation string
	genErr      error
	answer      string
	phraseErr   error
}

func (m *mockGeminiClient) GenerateSQL(ctx context.Context, question string) (services.SQLGeneration, error) {
	if m.genErr != nil {
		return services.SQLGeneration{}, m.genErr
	}
	return services.SQLGeneration{SQL: m.sql, Explanation: m.explanation}, nil
}

func (m *mockGeminiClient) PhraseAnswer(ctx context.Context, question string, resultsJSON string) (string, error) {
	if m.phraseErr != nil {
		return "", m.phraseErr
	}
	return m.answer, nil
}

func doAskRequest(t *testing.T, handler *handlers.AIHandler, body map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/ai/ask", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler.Ask(c)
	return w
}

func TestAIHandler_Ask_MissingQuestion(t *testing.T) {
	handler := handlers.NewAIHandler(&mockGeminiClient{})
	w := doAskRequest(t, handler, map[string]interface{}{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

func TestAIHandler_Ask_GenerationFailure(t *testing.T) {
	handler := handlers.NewAIHandler(&mockGeminiClient{genErr: errors.New("gemini unavailable")})
	w := doAskRequest(t, handler, map[string]interface{}{"question": "how many leads are open"})

	assert.Equal(t, http.StatusBadGateway, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "AI_GENERATION_FAILED", resp["code"])
}

func TestAIHandler_Ask_GuardrailRejectsNonSelect(t *testing.T) {
	handler := handlers.NewAIHandler(&mockGeminiClient{sql: "DELETE FROM leads WHERE id = '1'"})
	w := doAskRequest(t, handler, map[string]interface{}{"question": "delete all leads"})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "AI_GUARDRAIL_REJECTED", resp["code"])
}

func TestAIHandler_Ask_GuardrailRejectsMultiStatement(t *testing.T) {
	handler := handlers.NewAIHandler(&mockGeminiClient{sql: "SELECT 1; DROP TABLE leads;"})
	w := doAskRequest(t, handler, map[string]interface{}{"question": "anything"})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "AI_GUARDRAIL_REJECTED", resp["code"])
}

func TestAIHandler_Ask_GuardrailRejectsDisallowedTable(t *testing.T) {
	handler := handlers.NewAIHandler(&mockGeminiClient{sql: "SELECT * FROM agents_password_dump"})
	w := doAskRequest(t, handler, map[string]interface{}{"question": "show me agent passwords"})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "AI_GUARDRAIL_REJECTED", resp["code"])
}

func TestAIHandler_Ask_RejectsConfidentialQuestionBeforeCallingGemini(t *testing.T) {
	// sql/genErr are both left unset — if the handler called GenerateSQL at
	// all, mockGeminiClient would return an empty SQLGeneration{}, which
	// ValidateAIGeneratedSQL rejects as "empty query" with code
	// AI_GUARDRAIL_REJECTED. Getting AI_CONFIDENTIAL_DATA_REQUESTED instead
	// confirms the short-circuit happens before generation, not after.
	handler := handlers.NewAIHandler(&mockGeminiClient{})
	w := doAskRequest(t, handler, map[string]interface{}{"question": "what is the password for agent shankar"})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "AI_CONFIDENTIAL_DATA_REQUESTED", resp["code"])
}
