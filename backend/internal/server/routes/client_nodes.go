package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const maxClientNodeRequestSize = service.MaxClientNodeContentSize + 4096

type clientNodePolicyInput struct {
	Enabled       *bool    `json:"enabled"`
	DisabledNodes []string `json:"disabled_nodes"`
}

func readClientNodeJSON(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxClientNodeRequestSize)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		response.BadRequest(c, "invalid client node request")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.BadRequest(c, "only one JSON object is allowed")
		return false
	}
	return true
}

func clientNodePolicy(input clientNodePolicyInput) (service.ClientNodePolicy, bool) {
	if input.Enabled == nil || input.DisabledNodes == nil {
		return service.ClientNodePolicy{}, false
	}
	for _, name := range input.DisabledNodes {
		if name == "" {
			return service.ClientNodePolicy{}, false
		}
	}
	return service.ClientNodePolicy{Enabled: *input.Enabled, DisabledNodes: input.DisabledNodes}, true
}

// RegisterClientNodeRoutes keeps user reads and super-admin writes under their respective auth gates.
func RegisterClientNodeRoutes(v1 *gin.RouterGroup, jwtAuth middleware.JWTAuthMiddleware, adminAuth middleware.AdminAuthMiddleware, auditLog middleware.AuditLogMiddleware, settings *service.SettingService, panelRateLimiter *middleware.PanelRateLimiter) {
	client := v1.Group("/client", gin.HandlerFunc(jwtAuth))
	client.GET("/node-policy", func(c *gin.Context) {
		policy, err := settings.GetClientNodePolicy(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		c.JSON(http.StatusOK, policy)
	})
	client.GET("/node-subscription", func(c *gin.Context) {
		content, err := settings.GetClientNodeSubscription(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", []byte(content))
	})

	admin := v1.Group("/admin/client", gin.HandlerFunc(adminAuth), middleware.AdminOnly(), gin.HandlerFunc(auditLog), middleware.AdminComplianceGuard(settings))
	if panelRateLimiter != nil {
		admin.Use(panelRateLimiter.Global())
	}
	admin.GET("/node-policy", func(c *gin.Context) {
		policy, err := settings.GetClientNodePolicy(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, policy)
	})
	admin.PUT("/node-policy", func(c *gin.Context) {
		var input clientNodePolicyInput
		if !readClientNodeJSON(c, &input) {
			return
		}
		policy, ok := clientNodePolicy(input)
		if !ok {
			response.BadRequest(c, "enabled and disabled_nodes are required")
			return
		}
		updated, err := settings.SetClientNodePolicy(c.Request.Context(), policy)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, updated)
	})
	admin.GET("/node-subscription", func(c *gin.Context) {
		content, err := settings.GetClientNodeSubscription(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, gin.H{"content": content})
	})
	admin.PUT("/node-subscription", func(c *gin.Context) {
		var input struct {
			Content *string `json:"content"`
		}
		if !readClientNodeJSON(c, &input) {
			return
		}
		if input.Content == nil {
			response.BadRequest(c, "content is required")
			return
		}
		if err := settings.SetClientNodeSubscription(c.Request.Context(), *input.Content); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		response.Success(c, gin.H{"content": *input.Content})
	})
	admin.PUT("/nodes", func(c *gin.Context) {
		var input struct {
			Enabled       *bool    `json:"enabled"`
			DisabledNodes []string `json:"disabled_nodes"`
			Content       *string  `json:"content"`
		}
		if !readClientNodeJSON(c, &input) {
			return
		}
		policy, ok := clientNodePolicy(clientNodePolicyInput{Enabled: input.Enabled, DisabledNodes: input.DisabledNodes})
		if !ok || input.Content == nil {
			response.BadRequest(c, "enabled, disabled_nodes and content are required")
			return
		}
		if err := settings.SetClientNodes(c.Request.Context(), policy, *input.Content); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		response.Success(c, gin.H{"enabled": policy.Enabled, "disabled_nodes": policy.DisabledNodes, "content": *input.Content})
	})
}
