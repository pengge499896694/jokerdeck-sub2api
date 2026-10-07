package handler

import (
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSelfMonitorAdminOperationsDenyUsers(t *testing.T) {
	for name, action := range map[string]func(*ChannelMonitorUserHandler, *gin.Context){"run": (*ChannelMonitorUserHandler).SelfRun, "switch": (*ChannelMonitorUserHandler).SelfToggle} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(string(middleware.ContextKeyUserRole), "user")
			action(&ChannelMonitorUserHandler{}, c)
			require.Equal(t, 403, w.Code)
		})
	}
}

func TestSelfMonitorRunRequiresExplicitConfirmation(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(middleware.ContextKeyUserRole), "admin")
	c.Request = httptest.NewRequest("POST", "/self-channel-monitors/run", strings.NewReader(`{"id":1,"confirmed":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	(&ChannelMonitorUserHandler{}).SelfRun(c)
	require.Equal(t, 400, w.Code)
}
