package catalog

import (
	"time"

	"github.com/google/uuid"
)

type AcademicUnit struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Active      bool      `json:"active"`
	CourseCount int       `json:"course_count"`
}

type Branch struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	CourseCount int       `json:"course_count"`
}

type Course struct {
	ID                uuid.UUID      `json:"id"`
	Code              string         `json:"code"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Active            bool           `json:"active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ResourceCount     int            `json:"resource_count"`
	Branches          []Branch       `json:"branches"`
	AcademicUnits     []AcademicUnit `json:"academic_units"`
}
