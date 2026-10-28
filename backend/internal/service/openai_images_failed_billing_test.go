package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardImagesNoOutputNeverReturnsBillableResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name, contentType, response string
		stream                      bool
	}{
		{name: "empty images", response: `{"data":[],"usage":{"input_tokens":0,"output_tokens":0}}`},
		{name: "null image", response: `{"data":[null,{}, {"b64_json":""}]}`},
		{name: "business error with HTTP 200", response: `{"error":{"type":"server_error","message":"generation failed"}}`},
		{name: "rejection with HTTP 200", response: `{"error":{"type":"invalid_request_error","code":"content_policy_violation","message":"Blocked"}}`},
		{name: "html error with HTTP 200", contentType: "text/html", response: `<html>upstream unavailable</html>`},
		{name: "invalid JSON", response: `{"data":`},
		{name: "done without image", stream: true, response: "data: [DONE]\n\n"},
		{name: "usage without image", stream: true, response: "data: {\"usage\":{\"input_tokens\":0,\"output_tokens\":0}}\n\ndata: [DONE]\n\n"},
		{name: "stream fails", stream: true, response: "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"type\":\"server_error\",\"message\":\"Failed\"}}}\n\n"},
		{name: "stream error envelope", stream: true, response: "data: {\"error\":{\"type\":\"server_error\",\"message\":\"Failed\"}}\n\n"},
	}
	for _, enabled := range []bool{false, true} {
		for _, tt := range cases {
			t.Run(tt.name+map[bool]string{false: "/per-request", true: "/per-image"}[enabled], func(t *testing.T) {
				body := []byte(`{"model":"gpt-image-2","prompt":"draw a cat","n":3,"stream":` + map[bool]string{false: "false", true: "true"}[tt.stream] + `}`)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Set("api_key", &APIKey{Group: &Group{ImageCountBilling: enabled}})
				contentType := tt.contentType
				if contentType == "" {
					contentType = "application/json"
				}
				if tt.stream {
					contentType = "text/event-stream"
				}
				svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(tt.response))}}}
				parsed, err := svc.ParseOpenAIImagesRequest(c, body)
				require.NoError(t, err)
				account := &Account{ID: 6, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-key", "base_url": "https://images.example/v1"}}
				result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
				require.Error(t, err, "empty/error payloads must not count as successful generation")
				require.Nil(t, result, "the handler must not receive a result it can charge")
				if !tt.stream {
					require.NotContains(t, rec.Body.String(), `"data":[]`, "do not commit an empty successful response before failover")
				}
			})
		}
	}
}

func TestForwardImagesBillsOnlyActualOutputsEvenAfterStreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name, response  string
		stream, wantErr bool
	}{
		{name: "one image out of three", response: `{"data":[{"b64_json":"aGVsbG8="}]}`},
		{name: "one streamed image", stream: true, response: "data: {\"type\":\"image_generation.completed\",\"b64_json\":\"aGVsbG8=\"}\n\ndata: [DONE]\n\n"},
		{name: "completed image before late failure", stream: true, wantErr: true, response: "data: {\"type\":\"image_generation.completed\",\"b64_json\":\"aGVsbG8=\"}\n\ndata: {\"type\":\"error\",\"error\":{\"type\":\"server_error\",\"message\":\"next image failed\"}}\n\n"},
		{name: "JSON fallback despite stream content type", stream: true, response: `{"data":[{"b64_json":"aGVsbG8="}]}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-image-2","prompt":"draw","n":3,"stream":` + map[bool]string{false: "false", true: "true"}[tt.stream] + `}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			contentType := "application/json"
			if tt.stream {
				contentType = "text/event-stream"
			}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(tt.response))}}}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			account := &Account{ID: 6, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-key"}}
			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NotNil(t, result)
			require.Equal(t, 1, result.ImageCount, "a late failure must not erase an already completed image or charge missing images")
		})
	}
}
