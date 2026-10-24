package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractConversationQuestionAcrossProtocols(t *testing.T) {
	t.Run("anthropic text and ignores media", func(t *testing.T) {
		body := []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"What is in this picture?"},{"type":"image","source":{"data":"large-base64"}}]},{"role":"assistant","content":"not a question"}]}`)
		require.Equal(t, "What is in this picture?", extractConversationQuestion(body))
	})

	t.Run("responses input items", func(t *testing.T) {
		body := []byte(`{"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"First"},{"type":"input_image","image_url":"https://example.test/image.png"}]},{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ignore"}]}]}`)
		require.Equal(t, "First", extractConversationQuestion(body))
	})

	t.Run("gemini parts", func(t *testing.T) {
		body := []byte(`{"contents":[{"role":"user","parts":[{"text":"Explain this"},{"inline_data":{"mime_type":"image/png","data":"base64"}}]},{"role":"model","parts":[{"text":"not a question"}]}]}`)
		require.Equal(t, "Explain this", extractConversationQuestion(body))
	})
}

func TestExtractConversationAnswerAcrossProtocols(t *testing.T) {
	t.Run("openai json", func(t *testing.T) {
		answer, size, truncated := extractConversationAnswer([]byte(`{"choices":[{"message":{"content":[{"type":"text","text":"Hello"}]}}]}`), false)
		require.Equal(t, "Hello", answer)
		require.Equal(t, len("Hello"), size)
		require.False(t, truncated)
	})

	t.Run("anthropic json", func(t *testing.T) {
		answer, _, _ := extractConversationAnswer([]byte(`{"content":[{"type":"text","text":"Hello"},{"type":"image","source":{"data":"ignore"}}]}`), false)
		require.Equal(t, "Hello", answer)
	})

	t.Run("gemini json", func(t *testing.T) {
		answer, _, _ := extractConversationAnswer([]byte(`{"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}`), false)
		require.Equal(t, "Hello", answer)
	})

	t.Run("server-sent events prefer deltas over final snapshot", func(t *testing.T) {
		body := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\ndata: {\"choices\":[{\"message\":{\"content\":\"Hello\"}}]}\n\ndata: [DONE]\n")
		answer, _, truncated := extractConversationAnswer(body, false)
		require.Equal(t, "Hello", answer)
		require.False(t, truncated)
	})

	t.Run("records response buffer truncation", func(t *testing.T) {
		answer, size, truncated := extractConversationAnswer([]byte(`{"content":"partial"}`), true)
		require.Equal(t, "partial", answer)
		require.Equal(t, len("partial"), size)
		require.True(t, truncated)
	})
}

func TestConversationCaptureRouteSelection(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
		want   bool
	}{
		{method: "POST", path: "/v1/messages", want: true},
		{method: "POST", path: "/v1/chat/completions", want: true},
		{method: "POST", path: "/v1/responses", want: true},
		{method: "POST", path: "/v1beta/models/gemini-pro:streamGenerateContent", want: true},
		{method: "POST", path: "/v1/messages/count_tokens", want: false},
		{method: "POST", path: "/v1/images/generations", want: false},
		{method: "GET", path: "/v1/models", want: false},
	} {
		require.Equal(t, tc.want, isConversationCaptureRoute(tc.method, tc.path), "%s %s", tc.method, tc.path)
	}
}

func TestTruncateConversationTextPreservesUTF8(t *testing.T) {
	value, truncated := truncateConversationText("你好", 5)
	require.Equal(t, "你", value)
	require.True(t, truncated)
}
