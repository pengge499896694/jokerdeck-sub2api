package admin

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const maxClientNodeYAMLUploadSize = 1024 * 1024

func (h *SettingHandler) GetClientNodePolicy(c *gin.Context) {
	policy, err := h.settingService.GetGlobalClientNodePolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}

func (h *SettingHandler) UpdateClientNodePolicy(c *gin.Context) {
	var request service.ClientNodePolicy
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	policy, err := h.settingService.SetGlobalClientNodePolicy(c.Request.Context(), &request)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, policy)
}

func (h *SettingHandler) ListClientNodes(c *gin.Context) {
	nodes, err := h.settingService.ListClientNodeSubscriptions(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nodes)
}

func (h *SettingHandler) UploadClientNodeYAML(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxClientNodeYAMLUploadSize+(64<<10))
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择有效的 .yaml 或 .yml 文件")
		return
	}
	defer func() { _ = file.Close() }()
	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension != ".yaml" && extension != ".yml" {
		response.BadRequest(c, "文件扩展名必须是 .yaml 或 .yml")
		return
	}
	if header.Size > maxClientNodeYAMLUploadSize {
		response.BadRequest(c, "YAML 文件不能超过 1 MiB")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, maxClientNodeYAMLUploadSize+1))
	if err != nil {
		response.BadRequest(c, "读取 YAML 文件失败")
		return
	}
	if len(data) > maxClientNodeYAMLUploadSize {
		response.BadRequest(c, "YAML 文件不能超过 1 MiB")
		return
	}
	if err := h.settingService.SetUploadedClientNodeYAML(c.Request.Context(), data); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"uploaded": true})
}

func (h *SettingHandler) DeleteClientNodeYAML(c *gin.Context) {
	if err := h.settingService.DeleteUploadedClientNodeYAML(c.Request.Context()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SettingHandler) TestClientNodes(c *gin.Context) {
	var request struct {
		SubscriptionID string   `json:"subscription_id"`
		Names          []string `json:"names"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	results, err := h.settingService.TestClientNodeNodes(c.Request.Context(), request.SubscriptionID, request.Names)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, results)
}
