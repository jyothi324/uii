package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// aiChatModel is the single Gemini model used for both SQL generation and
// answer phrasing (per product decision — one fast/cheap model rather than
// a two-tier split).
//
// NOTE: gemini-2.5-flash and gemini-2.0-flash both return 404 ("no longer
// available") for API keys created after Google's cutover — verified
// directly against the generateContent endpoint. gemini-3.5-flash is the
// current working flash-tier model as of this writing; if this starts
// 404ing again in the future, check `GET /v1beta/models?key=...` for the
// currently available flash model and update this constant.
const aiChatModel = "gemini-3.5-flash"

const geminiDefaultBaseURL = "https://generativelanguage.googleapis.com"
const geminiGenerateContentURLFmt = "%s/v1beta/models/%s:generateContent?key=%s"

// SQLGeneration is the structured output of the SQL-generation step.
type SQLGeneration struct {
	SQL         string `json:"sql"`
	Explanation string `json:"explanation"`
}

// GeminiClient is the interface the AI handler depends on, so tests can
// substitute a mock instead of making real HTTP calls (consistent with the
// mocks/ package's testify/mock convention used elsewhere in this repo).
type GeminiClient interface {
	GenerateSQL(ctx context.Context, question string) (SQLGeneration, error)
	PhraseAnswer(ctx context.Context, question string, resultsJSON string) (string, error)
}

// HTTPGeminiClient is the real GeminiClient implementation: plain net/http
// + encoding/json, matching the existing hand-rolled REST style in
// inventory/services/gemini_service.go rather than pulling in an SDK.
type HTTPGeminiClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string // overridable in tests; defaults to geminiDefaultBaseURL
}

// NewHTTPGeminiClient reads GEMINI_API_KEY the same way
// inventory/services/gemini_service.go does (os.Getenv at construction
// time). Callers should treat an empty key as a startup configuration
// error, not something to silently retry around.
func NewHTTPGeminiClient() *HTTPGeminiClient {
	return &HTTPGeminiClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     os.Getenv("GEMINI_API_KEY"),
		baseURL:    geminiDefaultBaseURL,
	}
}

// NewHTTPGeminiClientForTest builds a client pointed at a custom base URL
// (e.g. an httptest.Server) instead of the real Gemini API. Used only by
// this package's tests.
func NewHTTPGeminiClientForTest(baseURL, apiKey string) *HTTPGeminiClient {
	return &HTTPGeminiClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}

// --- Gemini generateContent request/response shapes (subset used here) ---

type geminiPart struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

type geminiToolConfig struct {
	FunctionCallingConfig geminiFunctionCallingConfig `json:"functionCallingConfig"`
}

type geminiGenerateContentRequest struct {
	SystemInstruction *geminiContent    `json:"system_instruction,omitempty"`
	Contents          []geminiContent   `json:"contents"`
	Tools             []geminiTool      `json:"tools,omitempty"`
	ToolConfig        *geminiToolConfig `json:"toolConfig,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiGenerateContentResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// generateSQLTool is the single function declaration offered to the model
// for the SQL-generation step, forced via toolConfig mode "ANY" so the
// response is always a structured function call rather than free text.
var generateSQLTool = geminiTool{
	FunctionDeclarations: []geminiFunctionDeclaration{
		{
			Name:        "generate_sql",
			Description: "Produce a single read-only PostgreSQL SELECT statement that answers the user's question, plus a short explanation of what it does.",
			Parameters: map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"sql":         map[string]interface{}{"type": "STRING", "description": "The SELECT statement, no trailing semicolon required."},
					"explanation": map[string]interface{}{"type": "STRING", "description": "One sentence describing what the query returns."},
				},
				"required": []string{"sql", "explanation"},
			},
		},
	},
}

// GenerateSQL asks Gemini to translate a natural-language question into a
// SQL query via forced function-calling. The returned SQL is NOT yet
// validated — callers MUST pass it through ValidateAIGeneratedSQL before
// executing it.
func (c *HTTPGeminiClient) GenerateSQL(ctx context.Context, question string) (SQLGeneration, error) {
	reqBody := geminiGenerateContentRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: AISchemaContext}}},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: question}}},
		},
		Tools: []geminiTool{generateSQLTool},
		ToolConfig: &geminiToolConfig{
			FunctionCallingConfig: geminiFunctionCallingConfig{
				Mode:                 "ANY",
				AllowedFunctionNames: []string{"generate_sql"},
			},
		},
	}

	resp, err := c.call(ctx, reqBody)
	if err != nil {
		return SQLGeneration{}, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return SQLGeneration{}, fmt.Errorf("gemini returned no candidates for SQL generation")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.FunctionCall == nil || part.FunctionCall.Name != "generate_sql" {
		return SQLGeneration{}, fmt.Errorf("gemini did not return the expected generate_sql function call")
	}

	sql, _ := part.FunctionCall.Args["sql"].(string)
	explanation, _ := part.FunctionCall.Args["explanation"].(string)
	if sql == "" {
		return SQLGeneration{}, fmt.Errorf("gemini returned an empty sql argument")
	}

	return SQLGeneration{SQL: sql, Explanation: explanation}, nil
}

// PhraseAnswer asks Gemini to turn query results into a short natural
// language answer. Used only for multi-row/complex results — simple
// scalar aggregates are formatted in Go without an extra API call (see
// handlers/ai.handler.go).
func (c *HTTPGeminiClient) PhraseAnswer(ctx context.Context, question string, resultsJSON string) (string, error) {
	prompt := fmt.Sprintf(
		"The user asked: %q\n\nThe database returned this JSON result:\n%s\n\n"+
			"Answer the user's question in one or two natural, concise sentences using ONLY the "+
			"values literally present in the JSON above. Do not mention SQL, JSON, or the database. "+
			"Do not invent, guess, or infer any name, number, email, or other value that does not "+
			"appear verbatim in the JSON — if the JSON is empty or doesn't contain what's needed to "+
			"answer, say plainly that the data isn't available instead of guessing.",
		question, resultsJSON,
	)

	reqBody := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: prompt}}},
		},
	}

	resp, err := c.call(ctx, reqBody)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no candidates for answer phrasing")
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	if text == "" {
		return "", fmt.Errorf("gemini returned an empty text answer")
	}
	return text, nil
}

func (c *HTTPGeminiClient) call(ctx context.Context, reqBody geminiGenerateContentRequest) (*geminiGenerateContentResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = geminiDefaultBaseURL
	}
	url := fmt.Sprintf(geminiGenerateContentURLFmt, baseURL, aiChatModel, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result geminiGenerateContentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	return &result, nil
}
