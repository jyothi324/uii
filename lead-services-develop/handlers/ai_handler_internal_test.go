package handlers

import (
	"context"
	"testing"

	"github.com/tuliporg/agent-management-api/services"
)

// trackingGeminiClient records whether PhraseAnswer was called, so tests can
// assert Gemini is never invoked to "phrase" an answer from zero rows —
// that's exactly the case that produced a fabricated name/count/email in
// practice (see formatAnswer's comment).
type trackingGeminiClient struct {
	phraseAnswerCalled bool
	phraseAnswerReturn string
}

func (m *trackingGeminiClient) GenerateSQL(ctx context.Context, question string) (services.SQLGeneration, error) {
	return services.SQLGeneration{}, nil
}

func (m *trackingGeminiClient) PhraseAnswer(ctx context.Context, question string, resultsJSON string) (string, error) {
	m.phraseAnswerCalled = true
	return m.phraseAnswerReturn, nil
}

func TestFormatAnswer_EmptyResultsSkipsGeminiEntirely(t *testing.T) {
	mock := &trackingGeminiClient{phraseAnswerReturn: "this should never be returned"}
	handler := &AIHandler{gemini: mock}

	answer, err := handler.formatAnswer(context.Background(), "which customer has the most leads", []map[string]interface{}{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.phraseAnswerCalled {
		t.Error("expected PhraseAnswer NOT to be called for zero rows — this is the hallucination-prevention guarantee")
	}
	if answer != "I couldn't find any data matching that question." {
		t.Errorf("unexpected fixed message: %q", answer)
	}
}

func TestFormatAnswer_ScalarAggregateSkipsGemini(t *testing.T) {
	mock := &trackingGeminiClient{phraseAnswerReturn: "this should never be returned"}
	handler := &AIHandler{gemini: mock}

	results := []map[string]interface{}{{"count": 42}}
	answer, err := handler.formatAnswer(context.Background(), "how many leads", results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.phraseAnswerCalled {
		t.Error("expected PhraseAnswer NOT to be called for a single scalar row")
	}
	if answer != "The answer is 42." {
		t.Errorf("unexpected answer: %q", answer)
	}
}

func TestFormatAnswer_NonEmptyMultiRowCallsGemini(t *testing.T) {
	mock := &trackingGeminiClient{phraseAnswerReturn: "phrased answer"}
	handler := &AIHandler{gemini: mock}

	results := []map[string]interface{}{
		{"name": "Alice", "count": 3},
		{"name": "Bob", "count": 2},
	}
	answer, err := handler.formatAnswer(context.Background(), "list leads by agent", results)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !mock.phraseAnswerCalled {
		t.Error("expected PhraseAnswer to be called for a real multi-row result")
	}
	if answer != "phrased answer" {
		t.Errorf("unexpected answer: %q", answer)
	}
}
