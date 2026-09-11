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

type Handler struct { resourcesService *Service; store storage.ObjectStore; renderer Renderer }
func NewHandler(resourcesService *Service, store storage.ObjectStore, renderer Renderer) *Handler { return &Handler{resourcesService:resourcesService,store:store,renderer:renderer} }

func (h *Handler) HandleResourceDetail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r,"id")); if err != nil { http.NotFound(w,r); return }
	res, err := h.resourcesService.GetByID(r.Context(),id,true)
	if err != nil { if errors.Is(err,ErrNotFound){http.NotFound(w,r);return}; http.Error(w,"Failed to load resource",500);return }
	h.renderer.Render(w,"resource.html",map[string]interface{}{"Title":fmt.Sprintf("%s %s %s — %s",res.CourseCode,res.AcademicYearDisplay(),res.ExamType,res.CourseName),"Resource":res})
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) { h.streamFile(w,r,"attachment") }
func (h *Handler) HandlePreview(w http.ResponseWriter, r *http.Request) { h.streamFile(w,r,"inline") }

func (h *Handler) streamFile(w http.ResponseWriter,r *http.Request,disposition string){
	id,err:=uuid.Parse(chi.URLParam(r,"id"));if err!=nil{http.NotFound(w,r);return}
	res,err:=h.resourcesService.GetByID(r.Context(),id,true);if err!=nil{if errors.Is(err,ErrNotFound){http.NotFound(w,r);return};http.Error(w,"Failed to load resource",500);return}
	reader,err:=h.store.Get(r.Context(),res.StorageIdentifier);if err!=nil{http.Error(w,"Resource file is temporarily unavailable",http.StatusServiceUnavailable);return};defer reader.Close()
	name:=res.StorageFilename;if name==""{name=fmt.Sprintf("%s-%d-%s.pdf",res.CourseCode,res.AcademicYearStart,res.ExamType)}
	w.Header().Set("Content-Type","application/pdf")
	w.Header().Set("Content-Disposition",fmt.Sprintf(`%s; filename="%s"`,disposition,name))
	w.Header().Set("X-Content-Type-Options","nosniff")
	if res.FileSizeBytes>0{w.Header().Set("Content-Length",strconv.FormatInt(res.FileSizeBytes,10))}
	_,_=io.Copy(w,reader)
}

func (h *Handler) HandleSearch(w http.ResponseWriter,r *http.Request){
	q:=strings.TrimSpace(r.URL.Query().Get("q"));branch:=strings.TrimSpace(r.URL.Query().Get("branch"));course:=strings.TrimSpace(r.URL.Query().Get("course"));semester:=strings.TrimSpace(r.URL.Query().Get("semester"));examType:=strings.TrimSpace(r.URL.Query().Get("exam_type"));yearStr:=strings.TrimSpace(r.URL.Query().Get("year"));year,_:=strconv.Atoi(yearStr)
	page,_:=strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page")));if page<1{page=1};limit:=20;offset:=(page-1)*limit
	results,totalCount,err:=h.resourcesService.Search(r.Context(),FilterParams{Query:q,BranchCode:branch,CourseCode:course,Semester:semester,ExamType:examType,Year:year,Limit:limit,Offset:offset});if err!=nil{http.Error(w,"Search failed",500);return}
	totalPages:=(totalCount+limit-1)/limit;if totalPages<1{totalPages=1}
	data:=map[string]interface{}{"Title":"Search Papers — NITC Resource Vault","Query":q,"Branch":branch,"Course":course,"Semester":semester,"ExamType":examType,"Year":yearStr,"Results":results,"TotalCount":totalCount,"CurrentPage":page,"TotalPages":totalPages,"HasNext":page<totalPages,"HasPrev":page>1}
	if r.Header.Get("HX-Request")=="true"{h.renderer.RenderPartial(w,"search_results.html",data);return};h.renderer.Render(w,"search.html",data)
}
