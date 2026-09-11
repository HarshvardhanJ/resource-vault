package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/HarshvardhanJ/resource-vault/internal/storage"
)

type branchSeed struct {
	Code string
	Name string
}

type courseSeed struct {
	Code        string
	Name        string
	Description string
	Branches    []string
}

type resourceSeed struct {
	CourseCode        string
	ResourceTitle     string
	Semester          string
	AcademicYearStart int
	ExamType          string
	StorageIdentifier string
	StorageFilename   string
	StorageProvider   string
	FileSize          int64
	SHA256            string
}

func (db *DB) Seed(ctx context.Context, store storage.ObjectStore) error {
	branches := []branchSeed{
		{"BARCH", "Bachelor of Architecture (B.Arch.)"},
		{"BT", "Biotechnology"},
		{"CH", "Chemical Engineering"},
		{"CE", "Civil Engineering"},
		{"CSE", "Computer Science and Engineering"},
		{"EEE", "Electrical and Electronics Engineering"},
		{"ECE", "Electronics and Communication Engineering"},
		{"ENE", "Energy Engineering"},
		{"EP", "Engineering Physics"},
		{"HSS", "Humanities and Social Sciences"},
		{"MSE", "Materials Science and Engineering"},
		{"ME", "Mechanical Engineering"},
		{"PE", "Production Engineering"},
		{"ITEP", "4-year Integrated Teacher Education Programme (ITEP) B.Sc-B.Ed"},
	}

	branchIDs := make(map[string]uuid.UUID)
	for _, b := range branches {
		var id uuid.UUID
		err := db.Pool.QueryRow(ctx, `
			INSERT INTO branches (code, name, active)
			VALUES ($1, $2, true)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, b.Code, b.Name).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to seed branch %s: %w", b.Code, err)
		}
		branchIDs[b.Code] = id
	}

	courses := []courseSeed{
		{
			Code:        "CS201",
			Name:        "Data Structures",
			Description: "Abstract data types, arrays, linked lists, stacks, queues, trees, and graphs.",
			Branches:    []string{"CSE"},
		},
		{
			Code:        "CS202",
			Name:        "Discrete Mathematics",
			Description: "Set theory, propositional logic, relations, functions, and graph theory.",
			Branches:    []string{"CSE"},
		},
		{
			Code:        "CS301",
			Name:        "Theory of Computation",
			Description: "Automata theory, regular expressions, context-free grammars, Turing machines.",
			Branches:    []string{"CSE"},
		},
		{
			Code:        "CS303",
			Name:        "Operating Systems",
			Description: "Process management, concurrency, memory management, virtual memory, file systems.",
			Branches:    []string{"CSE"},
		},
		{
			Code:        "EC301",
			Name:        "Signals & Systems",
			Description: "Continuous and discrete-time signals, LTI systems, Fourier analysis, Laplace transform, z-transform.",
			Branches:    []string{"ECE", "EEE"},
		},
		{
			Code:        "EC201",
			Name:        "Digital Electronics",
			Description: "Boolean algebra, combinational and sequential logic circuits, counters, registers.",
			Branches:    []string{"ECE", "EEE", "CSE"},
		},
		{
			Code:        "EE201",
			Name:        "Electric Circuits & Networks",
			Description: "Network theorems, transient analysis, AC circuits, two-port networks.",
			Branches:    []string{"EEE"},
		},
		{
			Code:        "ME203",
			Name:        "Thermodynamics",
			Description: "Laws of thermodynamics, thermodynamic cycles, entropy, and properties of pure substances.",
			Branches:    []string{"ME", "CH"},
		},
		{
			Code:        "CE201",
			Name:        "Fluid Mechanics",
			Description: "Fluid statics, kinematics, dynamics, Navier-Stokes, pipe flow.",
			Branches:    []string{"CE", "ME"},
		},
		{
			Code:        "MA101",
			Name:        "Mathematics I",
			Description: "Calculus, multivariable functions, differential equations, linear algebra.",
			Branches:    []string{"CSE", "ECE", "EEE", "ME", "CE", "CH", "BT", "BARCH", "ENE", "EP", "MSE", "PE"},
		},
		{
			Code:        "AR101",
			Name:        "Architectural Design I",
			Description: "Fundamentals of design, space planning, graphic representation, and composition.",
			Branches:    []string{"BARCH"},
		},
		{
			Code:        "BT201",
			Name:        "Biochemistry & Microbiology",
			Description: "Cellular organization, microbial growth, metabolic pathways, and enzymology.",
			Branches:    []string{"BT"},
		},
		{
			Code:        "EN201",
			Name:        "Energy Resources and Conversion",
			Description: "Renewable energy, solar, wind, biomass, fuel cells, and energy efficiency.",
			Branches:    []string{"ENE", "ME", "EEE"},
		},
		{
			Code:        "EP201",
			Name:        "Quantum Mechanics for Engineers",
			Description: "Wave-particle duality, Schrödinger equation, harmonic oscillator, and semiconductors.",
			Branches:    []string{"EP"},
		},
		{
			Code:        "HS101",
			Name:        "Professional Communication & Ethics",
			Description: "Technical writing, presentation skills, engineering ethics, and intellectual property.",
			Branches:    []string{"HSS", "CSE", "ECE", "EEE", "ME", "CE", "CH", "BT", "BARCH", "ENE", "EP", "MSE", "PE", "ITEP"},
		},
		{
			Code:        "MS201",
			Name:        "Structure and Properties of Materials",
			Description: "Crystal structures, defects, phase diagrams, mechanical behavior of polymers, metals, and ceramics.",
			Branches:    []string{"MSE", "ME", "PE"},
		},
		{
			Code:        "PE201",
			Name:        "Manufacturing Processes",
			Description: "Casting, metal forming, machining, welding, and additive manufacturing.",
			Branches:    []string{"PE", "ME"},
		},
		{
			Code:        "ED101",
			Name:        "Foundations of Education",
			Description: "Philosophical, psychological, and sociological perspectives in contemporary education.",
			Branches:    []string{"ITEP"},
		},
	}

	courseIDs := make(map[string]uuid.UUID)
	for _, c := range courses {
		var id uuid.UUID
		err := db.Pool.QueryRow(ctx, `
			INSERT INTO courses (code, name, description, active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
			RETURNING id
		`, c.Code, c.Name, c.Description).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to seed course %s: %w", c.Code, err)
		}
		courseIDs[c.Code] = id

		for _, bCode := range c.Branches {
			bID, ok := branchIDs[bCode]
			if !ok {
				continue
			}
			_, err := db.Pool.Exec(ctx, `
				INSERT INTO course_branches (course_id, branch_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, id, bID)
			if err != nil {
				return fmt.Errorf("failed to map course %s to branch %s: %w", c.Code, bCode, err)
			}
		}
	}

	// Seed Admin User
	var adminID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO users (google_sub, email, display_name, role, status)
		VALUES ('seed-admin-001', 'admin@nitc.ac.in', 'NITC Archive Admin', 'admin', 'active')
		ON CONFLICT (google_sub) DO UPDATE SET role = 'admin'
		RETURNING id
	`).Scan(&adminID)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Ensure sample PDF exists in store
	sampleKey := "sample.pdf"
	fixturePath := filepath.Join("fixtures", "sample.pdf")
	if f, err := os.Open(fixturePath); err == nil {
		stat, _ := f.Stat()
		_ = store.Put(ctx, sampleKey, f, stat.Size(), "application/pdf")
		_ = f.Close()
	}

	// Sample PYQs
	samplePYQs := []resourceSeed{
		{
			CourseCode:        "CS201",
			ResourceTitle:     "Data Structures End-Semester Examination 2025",
			Semester:          "S3",
			AcademicYearStart: 2025,
			ExamType:          "ENDSEM",
			StorageIdentifier: sampleKey,
			StorageFilename:   "CS201-2025-ENDSEM.pdf",
			StorageProvider:   "local",
			FileSize:          607,
			SHA256:            "f95cee98461c390c4879c8f32a43040e3d1b9ace8bbbcbf8fb3b1b521b32b151",
		},
		{
			CourseCode:        "CS201",
			ResourceTitle:     "Data Structures Mid-Semester Examination 2024",
			Semester:          "S3",
			AcademicYearStart: 2024,
			ExamType:          "MIDSEM",
			StorageIdentifier: sampleKey,
			StorageFilename:   "CS201-2024-MIDSEM.pdf",
			StorageProvider:   "local",
			FileSize:          607,
			SHA256:            "f95cee98461c390c4879c8f32a43040e3d1b9ace8bbbcbf8fb3b1b521b32b151",
		},
		{
			CourseCode:        "EC301",
			ResourceTitle:     "Signals & Systems End-Semester Examination 2025",
			Semester:          "S5",
			AcademicYearStart: 2025,
			ExamType:          "ENDSEM",
			StorageIdentifier: sampleKey,
			StorageFilename:   "EC301-2025-ENDSEM.pdf",
			StorageProvider:   "local",
			FileSize:          607,
			SHA256:            "f95cee98461c390c4879c8f32a43040e3d1b9ace8bbbcbf8fb3b1b521b32b151",
		},
		{
			CourseCode:        "ME203",
			ResourceTitle:     "Thermodynamics End-Semester Examination 2024",
			Semester:          "S3",
			AcademicYearStart: 2024,
			ExamType:          "ENDSEM",
			StorageIdentifier: sampleKey,
			StorageFilename:   "ME203-2024-ENDSEM.pdf",
			StorageProvider:   "local",
			FileSize:          607,
			SHA256:            "f95cee98461c390c4879c8f32a43040e3d1b9ace8bbbcbf8fb3b1b521b32b151",
		},
		{
			CourseCode:        "MA101",
			ResourceTitle:     "Mathematics I End-Semester Examination 2025",
			Semester:          "S1",
			AcademicYearStart: 2025,
			ExamType:          "ENDSEM",
			StorageIdentifier: sampleKey,
			StorageFilename:   "MA101-2025-ENDSEM.pdf",
			StorageProvider:   "local",
			FileSize:          607,
			SHA256:            "f95cee98461c390c4879c8f32a43040e3d1b9ace8bbbcbf8fb3b1b521b32b151",
		},
	}

	for _, pyq := range samplePYQs {
		cID, ok := courseIDs[pyq.CourseCode]
		if !ok {
			continue
		}

		// Check if already seeded
		var exists bool
		_ = db.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM resources 
				WHERE course_id = $1 AND academic_year_start = $2 AND exam_type = $3
			)
		`, cID, pyq.AcademicYearStart, pyq.ExamType).Scan(&exists)

		if !exists {
			_, err = db.Pool.Exec(ctx, `
				INSERT INTO resources (
					course_id, uploader_id, resource_type, semester,
					academic_year_start, exam_type, title, file_size_bytes, sha256,
					storage_provider, storage_identifier, storage_filename,
					status, archive_status, scan_status, published_at, created_at, updated_at
				) VALUES (
					$1, $2, 'PYQ', $3,
					$4, $5, $6, $7, $8,
					$9, $10, $11,
					'PUBLISHED', 'UPLOADED', 'PASSED', $12, $12, $12
				)
			`, cID, adminID, pyq.Semester,
				pyq.AcademicYearStart, pyq.ExamType, pyq.ResourceTitle, pyq.FileSize, pyq.SHA256,
				pyq.StorageProvider, pyq.StorageIdentifier, pyq.StorageFilename,
				time.Now())
			if err != nil {
				return fmt.Errorf("failed to seed resource for course %s: %w", pyq.CourseCode, err)
			}
		}
	}

	return nil
}
