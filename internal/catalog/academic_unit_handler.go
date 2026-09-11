package catalog

import (
    "net/http"
)

type AcademicUnitHandler struct {
    repo     *AcademicUnitRepository
    renderer Renderer
}

func NewAcademicUnitHandler(repo *AcademicUnitRepository, renderer Renderer) *AcademicUnitHandler {
    return &AcademicUnitHandler{repo: repo, renderer: renderer}
}

// HandleList renders the canonical academic-unit listing used by the public catalog.
func (h *AcademicUnitHandler) HandleList(w http.ResponseWriter, r *http.Request) {
    units, err := h.repo.ListActive(r.Context())
    if err != nil {
        http.Error(w, "Failed to load academic units", http.StatusInternalServerError)
        return
    }
    h.renderer.Render(w, "academic_units.html", map[string]interface{}{
        "Title": "Academic Units — NITC Resource Vault",
        "AcademicUnits": units,
    })
}
