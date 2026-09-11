package catalog

import (
    "errors"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/HarshvardhanJ/resource-vault/internal/resources"
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
    return &Handler{catalogService: catalogService, resourcesService: resourcesService, renderer: renderer}
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
    branches, err := h.catalogService.ListBranches(r.Context())
    if err != nil { http.Error(w, "Failed to load branches", http.StatusInternalServerError); return }
    recent, err := h.resourcesService.ListRecent(r.Context(), 10)
    if err != nil { recent = nil }
    totalCourses := 0
    for _, b := range branches { totalCourses += b.CourseCount }
    data := map[string]interface{}{
        "Title": "NITC Resource Vault — Public Academic Archive",
        "Branches": branches,
        "RecentPYQs": recent,
        "TotalCourses": totalCourses,
    }
    h.renderer.Render(w, "home.html", data)
}

func (h *Handler) HandleBranch(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "branch")
    if code == "" { http.NotFound(w, r); return }
    branch, err := h.catalogService.GetBranch(r.Context(), code)
    if err != nil {
        if errors.Is(err, ErrNotFound) { http.NotFound(w, r); return }
        http.Error(w, "Failed to load academic unit", http.StatusInternalServerError); return
    }
    courses, err := h.catalogService.ListCoursesByBranch(r.Context(), branch.Code)
    if err != nil { http.Error(w, "Failed to load courses", http.StatusInternalServerError); return }
    h.renderer.Render(w, "branch.html", map[string]interface{}{
        "Title": branch.Code + " — " + branch.Name,
        "Branch": branch,
        "Courses": courses,
    })
}

func (h *Handler) HandleCourse(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "course")
    if code == "" { http.NotFound(w, r); return }
    course, err := h.catalogService.GetCourse(r.Context(), code)
    if err != nil {
        if errors.Is(err, ErrNotFound) { http.NotFound(w, r); return }
        http.Error(w, "Failed to load course", http.StatusInternalServerError); return
    }
    examType := strings.TrimSpace(r.URL.Query().Get("exam_type"))
    semester := strings.TrimSpace(r.URL.Query().Get("semester"))
    yearStr := strings.TrimSpace(r.URL.Query().Get("year"))
    year, _ := strconv.Atoi(yearStr)
    resList, err := h.resourcesService.ListByCourse(r.Context(), course.ID, examType, year, semester)
    if err != nil { http.Error(w, "Failed to load course papers", http.StatusInternalServerError); return }
    h.renderer.Render(w, "course.html", map[string]interface{}{
        "Title": course.Code + " — " + course.Name,
        "Course": course,
        "Resources": resList,
        "SelectedExamType": examType,
        "SelectedSemester": semester,
        "SelectedYear": yearStr,
    })
}

func (h *Handler) HandleEditCourse(w http.ResponseWriter, r *http.Request) {
    courseCode := chi.URLParam(r, "course")
    course, err := h.catalogService.GetCourse(r.Context(), courseCode)
    if err != nil {
        if errors.Is(err, ErrNotFound) { http.NotFound(w, r); return }
        http.Error(w, "Failed to load course", http.StatusInternalServerError); return
    }
    allBranches, err := h.catalogService.ListBranches(r.Context())
    if err != nil { http.Error(w, "Failed to load academic units", http.StatusInternalServerError); return }
    mapped := map[string]bool{}
    for _, b := range course.Branches { mapped[b.Code] = true }
    h.renderer.Render(w, "course_edit.html", map[string]interface{}{
        "Title": "Edit " + course.Code + " — NITC Resource Vault",
        "Course": course, "Branches": allBranches, "MappedBranchMap": mapped,
    })
}

func (h *Handler) HandleUpdateCourse(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "course")
    if err := r.ParseForm(); err != nil { http.Error(w, "Invalid form submission", http.StatusBadRequest); return }
    updated, err := h.catalogService.UpdateCourse(r.Context(), code, strings.TrimSpace(r.FormValue("code")), strings.TrimSpace(r.FormValue("name")), strings.TrimSpace(r.FormValue("description")), r.Form["branches"])
    if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
    http.Redirect(w, r, "/courses/"+updated.Code, http.StatusSeeOther)
}

func (h *Handler) HandleNewCourse(w http.ResponseWriter, r *http.Request) {
    allBranches, err := h.catalogService.ListBranches(r.Context())
    if err != nil { http.Error(w, "Failed to load academic units", http.StatusInternalServerError); return }
    preselect := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("branch")))
    mapped := map[string]bool{}
    if preselect != "" { mapped[preselect] = true }
    h.renderer.Render(w, "course_new.html", map[string]interface{}{
        "Title": "Add New Course — NITC Resource Vault",
        "Branches": allBranches, "MappedBranchMap": mapped, "Preselected": preselect,
    })
}

func (h *Handler) HandleCreateCourse(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil { http.Error(w, "Invalid form submission", http.StatusBadRequest); return }
    created, err := h.catalogService.CreateCourse(r.Context(), strings.TrimSpace(r.FormValue("code")), strings.TrimSpace(r.FormValue("name")), strings.TrimSpace(r.FormValue("description")), r.Form["branches"])
    if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
    http.Redirect(w, r, "/courses/"+created.Code, http.StatusSeeOther)
}
