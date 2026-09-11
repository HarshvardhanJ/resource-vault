package resources

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ResourceStatus string

const (
	StatusPendingReview ResourceStatus = "PENDING_REVIEW"
	StatusPublished     ResourceStatus = "PUBLISHED"
	StatusRejected      ResourceStatus = "REJECTED"
	StatusRemoved       ResourceStatus = "REMOVED"
)

type ExamType string

const (
	ExamMidsem ExamType = "MIDSEM"
	ExamEndsem ExamType = "ENDSEM"
	ExamQuiz   ExamType = "QUIZ"
	ExamOther  ExamType = "OTHER"
)

type Resource struct {
	ID                uuid.UUID      `json:"id"`
	CourseID          uuid.UUID      `json:"course_id"`
	UploaderID        *uuid.UUID     `json:"uploader_id"`
	ResourceType      string         `json:"resource_type"`
	Semester          string         `json:"semester"`
	AcademicYearStart int            `json:"academic_year_start"`
	ExamType          string         `json:"exam_type"`
	Title             string         `json:"title"`
	FileSizeBytes     int64          `json:"file_size_bytes"`
	SHA256            string         `json:"sha256"`
	StorageProvider   string         `json:"storage_provider"`
	StorageIdentifier string         `json:"storage_identifier"`
	StorageFilename   string         `json:"storage_filename"`
	Status            ResourceStatus `json:"status"`
	ArchiveStatus     string         `json:"archive_status"`
	ScanStatus        string         `json:"scan_status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	PublishedAt       *time.Time     `json:"published_at"`
	RemovedAt         *time.Time     `json:"removed_at"`

	// Enriched fields
	CourseCode  string   `json:"course_code"`
	CourseName  string   `json:"course_name"`
	BranchCodes []string `json:"branch_codes"`
}

func (r *Resource) AcademicYearDisplay() string {
	if r.AcademicYearStart <= 0 {
		return ""
	}
	nextYear := (r.AcademicYearStart + 1) % 100
	return fmt.Sprintf("%d-%02d", r.AcademicYearStart, nextYear)
}

func (r *Resource) FileSizeDisplay() string {
	const unit = 1024
	if r.FileSizeBytes < unit {
		return fmt.Sprintf("%d B", r.FileSizeBytes)
	}
	div, exp := int64(unit), 0
	for n := r.FileSizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(r.FileSizeBytes)/float64(div), units[exp])
}

type FilterParams struct {
	CourseCode string
	BranchCode string
	Semester   string
	ExamType   string
	Year       int
	Query      string
	Limit      int
	Offset     int
}
