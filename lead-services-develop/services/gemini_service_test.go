package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPGeminiClient_GenerateSQL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body geminiGenerateContentRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ToolConfig == nil || body.ToolConfig.FunctionCallingConfig.Mode != "ANY" {
			t.Errorf("expected forced function calling mode ANY, got %+v", body.ToolConfig)
		}

		resp := geminiGenerateContentResponse{
			Candidates: []geminiCandidate{
				{
					Content: geminiContent{
						Parts: []geminiPart{
							{
								FunctionCall: &geminiFunctionCall{
									Name: "generate_sql",
									Args: map[string]interface{}{
										"sql":         "SELECT COUNT(*) AS count FROM leads WHERE deleted_at IS NULL",
										"explanation": "Counts non-deleted leads.",
									},
								},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPGeminiClientForTest(server.URL, "test-key")
	got, err := client.GenerateSQL(context.Background(), "how many leads are open")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(got.SQL, "SELECT") {
		t.Errorf("expected a SELECT statement, got %q", got.SQL)
	}
	if got.Explanation == "" {
		t.Error("expected a non-empty explanation")
	}
}

func TestHTTPGeminiClient_GenerateSQL_NoFunctionCallReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiGenerateContentResponse{
			Candidates: []geminiCandidate{
				{Content: geminiContent{Parts: []geminiPart{{Text: "I refuse to answer."}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPGeminiClientForTest(server.URL, "test-key")
	if _, err := client.GenerateSQL(context.Background(), "how many leads are open"); err == nil {
		t.Error("expected an error when Gemini doesn't return the expected function call")
	}
}

func TestHTTPGeminiClient_GenerateSQL_MissingAPIKey(t *testing.T) {
	client := NewHTTPGeminiClientForTest("http://example.invalid", "")
	if _, err := client.GenerateSQL(context.Background(), "anything"); err == nil {
		t.Error("expected an error when GEMINI_API_KEY is not configured")
	}
}

func TestHTTPGeminiClient_GenerateSQL_UpstreamErrorSurfaced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	client := NewHTTPGeminiClientForTest(server.URL, "test-key")
	if _, err := client.GenerateSQL(context.Background(), "anything"); err == nil {
		t.Error("expected upstream error status to surface as an error")
	}
}

func TestHTTPGeminiClient_PhraseAnswer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiGenerateContentResponse{
			Candidates: []geminiCandidate{
				{Content: geminiContent{Parts: []geminiPart{{Text: "There were 3 leads created yesterday."}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPGeminiClientForTest(server.URL, "test-key")
	answer, err := client.PhraseAnswer(context.Background(), "what leads were created yesterday", `[{"id":"1"},{"id":"2"},{"id":"3"}]`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if answer == "" {
		t.Error("expected a non-empty phrased answer")
	}
}
