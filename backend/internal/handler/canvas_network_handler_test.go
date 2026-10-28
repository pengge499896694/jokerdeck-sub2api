package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCanvasRelayTargetAllowsConfiguredHostAndSubdomain(t *testing.T) {
	allowed := []string{"github.com", "raw.githubusercontent.com"}
	for _, raw := range []string{"https://github.com/a.png?x=1", "https://gist.github.com/a", "https://raw.githubusercontent.com/u/r/main/a.png"} {
		target, err := parseCanvasRelayTarget(raw, allowed)
		require.NoError(t, err)
		require.NotNil(t, target)
	}
}

func TestParseCanvasRelayTargetRejectsSSRFAndUnconfiguredHosts(t *testing.T) {
	allowed := []string{"github.com"}
	for _, raw := range []string{"http://127.0.0.1/a", "http://169.254.169.254/latest", "https://evil.example/a", "https://user:pass@github.com/a", "javascript:alert(1)"} {
		target, err := parseCanvasRelayTarget(raw, allowed)
		require.Error(t, err)
		require.Nil(t, target)
	}
}
