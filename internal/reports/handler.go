package reports

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Renderer interface {
	Render(w http.ResponseWriter, name string, data interface{})
	RenderPartial(w http.ResponseWriter, name string, data interface{})
}

type Handler struct {
	pool     *pgxpool.Pool
	renderer Renderer
}

func NewHandler(pool *pgxpool.Pool, renderer Renderer) *Handler {
	return &Handler{pool: pool, renderer: renderer}
}

func (h *Handler) HandleNewReport(w http.ResponseWriter, r *http.Request) {
	resourceID := r.URL.Query().Get("resource")
	data := map[string]interface{}{
		"Title":      "Report Resource",
		"ResourceID": resourceID,
		"Submitted":  false,
	}
	h.renderer.Render(w, "report.html", data)
}

func (h *Handler) HandleSubmitReport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	resIDStr := strings.TrimSpace(r.FormValue("resource_id"))
	category := strings.TrimSpace(r.FormValue("category"))
	message := strings.TrimSpace(r.FormValue("message"))

	resID, err := uuid.Parse(resIDStr)
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	if category == "" {
		category = "OTHER"
	}

	_, err = h.pool.Exec(r.Context(), `
		INSERT INTO reports (resource_id, category, message, status)
		VALUES ($1, $2, $3, 'PENDING')
	`, resID, category, message)
	if err != nil {
		http.Error(w, "Failed to submit report", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":     "Report Submitted",
		"Submitted": true,
	}
	h.renderer.Render(w, "report.html", data)
}

func (h *Handler) HandleContributeInfo(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Contribute — NITC PYQ Archive",
	}
	h.renderer.Render(w, "contribute.html", data)
}
