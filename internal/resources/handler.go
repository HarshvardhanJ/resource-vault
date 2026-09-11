package resources

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/HarshvardhanJ/resource-vault/internal/storage"
)

type Renderer interface {
	Render(w http.ResponseWriter, name string, data interface{})
	RenderPartial(w http.ResponseWriter, name string, data interface{})
}

type Handler struct {
	resourcesService *Service
	store            storage.ObjectStore
	renderer         Renderer
}

func NewHandler(resourcesService *Service, store storage.ObjectStore, renderer Renderer) *Handler {
	return &Handler{
		resourcesService: resourcesService,
		store:            store,
		renderer:         renderer,
	}
}

func (h *Handler) HandleResourceDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Guest route: only published
	res, err := h.resourcesService.GetByID(r.Context(), id, true)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load resource", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    fmt.Sprintf("%s %s %s — %s", res.CourseCode, res.AcademicYearDisplay(), res.ExamType, res.CourseName),
		"Resource": res,
	}

	h.renderer.Render(w, "resource.html", data)
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	res, err := h.resourcesService.GetByID(r.Context(), id, true)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load resource", http.StatusInternalServerError)
		return
	}

	reader, err := h.store.Get(r.Context(), res.StorageIdentifier)
	if err != nil {
		http.Error(w, "Failed to retrieve resource file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	downloadFilename := res.StorageFilename
	if downloadFilename == "" {
		downloadFilename = fmt.Sprintf("%s-%d-%s.pdf", res.CourseCode, res.AcademicYearStart, res.ExamType)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadFilename))
	if res.FileSizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(res.FileSizeBytes, 10))
	}

	_, _ = io.Copy(w, reader)
}

func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	res, err := h.resourcesService.GetByID(r.Context(), id, true)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load resource", http.StatusInternalServerError)
		return
	}

	reader, err := h.store.Get(r.Context(), res.StorageIdentifier)
	if err != nil {
		http.Error(w, "Failed to retrieve preview file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, res.StorageFilename))
	if res.FileSizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(res.FileSizeBytes, 10))
	}

	_, _ = io.Copy(w, reader)
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	branch := strings.TrimSpace(r.URL.Query().Get("branch"))
	course := strings.TrimSpace(r.URL.Query().Get("course"))
	semester := strings.TrimSpace(r.URL.Query().Get("semester"))
	examType := strings.TrimSpace(r.URL.Query().Get("exam_type"))
	yearStr := strings.TrimSpace(r.URL.Query().Get("year"))
	year, _ := strconv.Atoi(yearStr)

	pageStr := strings.TrimSpace(r.URL.Query().Get("page"))
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	params := FilterParams{
		Query:      q,
		BranchCode: branch,
		CourseCode: course,
		Semester:   semester,
		ExamType:   examType,
		Year:       year,
		Limit:      limit,
		Offset:     offset,
	}

	results, totalCount, err := h.resourcesService.Search(r.Context(), params)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	totalPages := (totalCount + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	data := map[string]interface{}{
		"Title":            "Search Papers — NITC PYQ Archive",
		"Query":            q,
		"Branch":           branch,
		"Course":           course,
		"Semester":         semester,
		"ExamType":         examType,
		"Year":             yearStr,
		"Results":          results,
		"TotalCount":       totalCount,
		"CurrentPage":      page,
		"TotalPages":       totalPages,
		"HasNext":          page < totalPages,
		"HasPrev":          page > 1,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		h.renderer.RenderPartial(w, "search_results.html", data)
		return
	}

	h.renderer.Render(w, "search.html", data)
}
