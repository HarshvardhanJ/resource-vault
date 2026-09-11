package catalog

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nitc-pyq-archive/archive/internal/resources"
)

type Renderer interface {
	Render(w http.ResponseWriter, name string, data interface{})
	RenderPartial(w http.ResponseWriter, name string, data interface{})
}

type Handler struct {
	catalogService   *Service
	resourcesService *resources.Service
	renderer         Renderer
}

func NewHandler(catalogService *Service, resourcesService *resources.Service, renderer Renderer) *Handler {
	return &Handler{
		catalogService:   catalogService,
		resourcesService: resourcesService,
		renderer:         renderer,
	}
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	branches, err := h.catalogService.ListBranches(r.Context())
	if err != nil {
		http.Error(w, "Failed to load branches", http.StatusInternalServerError)
		return
	}

	recent, err := h.resourcesService.ListRecent(r.Context(), 10)
	if err != nil {
		recent = nil
	}

	totalCourses := 0
	for _, b := range branches {
		totalCourses += b.CourseCount
	}

	data := map[string]interface{}{
		"Title":        "NITC PYQ Archive — Public Academic Archive",
		"Branches":     branches,
		"RecentPYQs":   recent,
		"TotalCourses": totalCourses,
	}

	h.renderer.Render(w, "home.html", data)
}

func (h *Handler) HandleBranch(w http.ResponseWriter, r *http.Request) {
	branchCode := chi.URLParam(r, "branch")
	if branchCode == "" {
		http.NotFound(w, r)
		return
	}

	branch, err := h.catalogService.GetBranch(r.Context(), branchCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load branch", http.StatusInternalServerError)
		return
	}

	courses, err := h.catalogService.ListCoursesByBranch(r.Context(), branch.Code)
	if err != nil {
		http.Error(w, "Failed to load branch courses", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   branch.Code + " — " + branch.Name,
		"Branch":  branch,
		"Courses": courses,
	}

	h.renderer.Render(w, "branch.html", data)
}

func (h *Handler) HandleCourse(w http.ResponseWriter, r *http.Request) {
	courseCode := chi.URLParam(r, "course")
	if courseCode == "" {
		http.NotFound(w, r)
		return
	}

	course, err := h.catalogService.GetCourse(r.Context(), courseCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to load course", http.StatusInternalServerError)
		return
	}

	examType := strings.TrimSpace(r.URL.Query().Get("exam_type"))
	semester := strings.TrimSpace(r.URL.Query().Get("semester"))
	yearStr := strings.TrimSpace(r.URL.Query().Get("year"))
	year, _ := strconv.Atoi(yearStr)

	resList, err := h.resourcesService.ListByCourse(r.Context(), course.ID, examType, year, semester)
	if err != nil {
		http.Error(w, "Failed to load course papers", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":            course.Code + " — " + course.Name,
		"Course":           course,
		"Resources":        resList,
		"SelectedExamType": examType,
		"SelectedSemester": semester,
		"SelectedYear":     yearStr,
	}

	h.renderer.Render(w, "course.html", data)
}
