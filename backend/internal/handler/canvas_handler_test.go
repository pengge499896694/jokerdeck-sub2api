package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func canvasTestRouter(h *CanvasHandler, user int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: user})
		c.Next()
	})
	r.GET("/canvas/:id", h.Get)
	r.PUT("/canvas/:id", h.Save)
	r.POST("/canvas/:id/assets", h.Upload)
	r.GET("/canvas/:id/assets/:asset_id", h.Asset)
	return r
}

func TestCanvasDocumentBounds(t *testing.T) {
	require.True(t, validCanvasDocument(json.RawMessage(`{"nodes":[],"edges":[],"viewport":{"x":0,"y":0,"zoom":1}}`)))
	for _, raw := range []string{`{}`, `{"nodes":[],"edges":[],"viewport":{"zoom":0}}`, `{"nodes":[],"edges":[],"viewport":{"zoom":5}}`} {
		require.False(t, validCanvasDocument(json.RawMessage(raw)))
	}
}

func TestCanvasGetScopesByOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	id := uuid.NewString()
	mock.ExpectQuery("SELECT title, document, revision FROM canvas_boards WHERE id=\\$1 AND user_id=\\$2").WithArgs(uuid.MustParse(id), int64(42)).WillReturnRows(sqlmock.NewRows([]string{"title", "document", "revision"}).AddRow("Board", []byte(`{"nodes":[],"edges":[],"viewport":{"x":0,"y":0,"zoom":1}}`), 2))
	w := httptest.NewRecorder()
	canvasTestRouter(NewCanvasHandler(db), 42).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/canvas/"+id, nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"revision":2`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCanvasSaveConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	id := uuid.NewString()
	mock.ExpectQuery("UPDATE canvas_boards SET").WithArgs("Board", sqlmock.AnyArg(), uuid.MustParse(id), int64(42), int64(3)).WillReturnRows(sqlmock.NewRows([]string{"revision"}))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(uuid.MustParse(id), int64(42)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	body := `{"title":"Board","revision":3,"document":{"nodes":[],"edges":[],"viewport":{"x":0,"y":0,"zoom":1}}}`
	w := httptest.NewRecorder()
	canvasTestRouter(NewCanvasHandler(db), 42).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/canvas/"+id, bytes.NewBufferString(body)))
	require.Equal(t, http.StatusConflict, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCanvasUploadEnforcesQuota(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 100, A: 255})
	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, img))
	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	part, err := writer.CreateFormFile("image", "reference.png")
	require.NoError(t, err)
	_, err = part.Write(pngBytes.Bytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users").WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(canvasQuotaLimit))
	mock.ExpectRollback()
	req := httptest.NewRequest(http.MethodPost, "/canvas/"+uuid.NewString()+"/assets", &form)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	canvasTestRouter(NewCanvasHandler(db), 42).ServeHTTP(w, req)
	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
