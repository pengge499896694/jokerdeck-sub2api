package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	canvasDocumentLimit = 2 << 20
	canvasAssetLimit    = 8 << 20
	canvasQuotaLimit    = 100 << 20
)

type CanvasHandler struct{ db *sql.DB }

func NewCanvasHandler(db *sql.DB) *CanvasHandler { return &CanvasHandler{db: db} }

type canvasDocument struct {
	Nodes    []json.RawMessage `json:"nodes"`
	Edges    []json.RawMessage `json:"edges"`
	Viewport struct {
		X    float64 `json:"x"`
		Y    float64 `json:"y"`
		Zoom float64 `json:"zoom"`
	} `json:"viewport"`
}

func validCanvasDocument(raw json.RawMessage) bool {
	if len(raw) == 0 || len(raw) > canvasDocumentLimit {
		return false
	}
	var doc canvasDocument
	if err := json.Unmarshal(raw, &doc); err != nil || doc.Nodes == nil || doc.Edges == nil || len(doc.Nodes) > 300 || len(doc.Edges) > 600 {
		return false
	}
	return doc.Viewport.Zoom > 0 && doc.Viewport.Zoom <= 4
}

func canvasOwner(c *gin.Context) (int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return 0, false
	}
	return subject.UserID, true
}

func canvasFailure(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canvas not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Canvas storage unavailable"})
}

func (h *CanvasHandler) List(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	rows, err := h.db.QueryContext(c, `SELECT id, title, revision, updated_at FROM canvas_boards WHERE user_id=$1 ORDER BY updated_at DESC LIMIT 100`, user)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id, title string
		var revision int64
		var updated interface{}
		if err := rows.Scan(&id, &title, &revision, &updated); err != nil {
			canvasFailure(c, err)
			return
		}
		items = append(items, gin.H{"id": id, "title": title, "revision": revision, "updated_at": updated})
	}
	if err := rows.Err(); err != nil {
		canvasFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *CanvasHandler) Create(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	var input struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid title"})
		return
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must be 1-120 characters"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.ExecContext(c, `INSERT INTO canvas_boards (id,user_id,title) VALUES ($1,$2,$3)`, id, user, title)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "title": title, "revision": 0})
}

func (h *CanvasHandler) Get(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	var title string
	var revision int64
	var document json.RawMessage
	err = h.db.QueryRowContext(c, `SELECT title, document, revision FROM canvas_boards WHERE id=$1 AND user_id=$2`, id, user).Scan(&title, &document, &revision)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "title": title, "document": document, "revision": revision})
}

func (h *CanvasHandler) Save(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, canvasDocumentLimit+1024)
	var input struct {
		Title    string          `json:"title"`
		Document json.RawMessage `json:"document"`
		Revision int64           `json:"revision"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || !validCanvasDocument(input.Document) || input.Revision < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or oversized canvas"})
		return
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid title"})
		return
	}
	var revision int64
	err = h.db.QueryRowContext(c, `UPDATE canvas_boards SET title=$1, document=$2, revision=revision+1, updated_at=NOW() WHERE id=$3 AND user_id=$4 AND revision=$5 RETURNING revision`, title, input.Document, id, user, input.Revision).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if e := h.db.QueryRowContext(c, `SELECT EXISTS(SELECT 1 FROM canvas_boards WHERE id=$1 AND user_id=$2)`, id, user).Scan(&exists); e != nil {
			canvasFailure(c, e)
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Canvas changed in another tab. Reload before saving."})
			return
		}
	}
	if err != nil {
		canvasFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"revision": revision})
}

func (h *CanvasHandler) Delete(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	result, err := h.db.ExecContext(c, `DELETE FROM canvas_boards WHERE id=$1 AND user_id=$2`, id, user)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		canvasFailure(c, err)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CanvasHandler) Upload(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	board, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, canvasAssetLimit+(1<<20))
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image"})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, canvasAssetLimit+1))
	if err != nil || len(data) == 0 || len(data) > canvasAssetLimit {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Image limit is 8 MB"})
		return
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg" && format != "webp") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PNG, JPEG and WebP are supported"})
		return
	}
	mime := "image/" + format
	if format == "jpeg" {
		mime = "image/jpeg"
	}
	tx, err := h.db.BeginTx(c, nil)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	defer tx.Rollback()
	// Serialize a user's uploads so concurrent requests cannot bypass the quota.
	var locked int64
	if err = tx.QueryRowContext(c, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, user).Scan(&locked); err != nil {
		canvasFailure(c, err)
		return
	}
	var total int64
	err = tx.QueryRowContext(c, `SELECT COALESCE(SUM(size_bytes),0) FROM canvas_assets WHERE user_id=$1`, user).Scan(&total)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	if total+int64(len(data)) > canvasQuotaLimit {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Canvas image quota is 100 MB"})
		return
	}
	id := uuid.NewString()
	err = tx.QueryRowContext(c, `INSERT INTO canvas_assets (id,board_id,user_id,content_type,data,size_bytes) SELECT $1,id,$2,$3,$4,$5 FROM canvas_boards WHERE id=$6 AND user_id=$2 RETURNING id`, id, user, mime, data, len(data), board).Scan(&id)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	if err = tx.Commit(); err != nil {
		canvasFailure(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *CanvasHandler) Asset(c *gin.Context) {
	user, ok := canvasOwner(c)
	if !ok {
		return
	}
	board, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	id, err := uuid.Parse(c.Param("asset_id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	var mime string
	var data []byte
	err = h.db.QueryRowContext(c, `SELECT content_type,data FROM canvas_assets WHERE id=$1 AND board_id=$2 AND user_id=$3`, id, board, user).Scan(&mime, &data)
	if err != nil {
		canvasFailure(c, err)
		return
	}
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, mime, data)
}
