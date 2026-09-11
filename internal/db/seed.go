package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/HarshvardhanJ/resource-vault/internal/storage"
)

type branchSeed struct { Code, Name string }
type courseSeed struct { Code, Name, Description string; Branches []string }

func (db *DB) Seed(ctx context.Context, store storage.ObjectStore) error {
	units := []branchSeed{
		{"BARCH", "B. Arch"}, {"BT", "Biotechnology"}, {"CH", "Chemical Engineering"}, {"CE", "Civil Engineering"},
		{"CSE", "Computer Science and Engineering"}, {"EEE", "Electrical and Electronics Engineering"},
		{"ECE", "Electronics and Communication Engineering"}, {"ENE", "Energy Engineering"}, {"EP", "Engineering Physics"},
		{"HSS", "Humanities and Social Sciences"}, {"MSE", "Materials Science and Engineering"}, {"ME", "Mechanical Engineering"},
		{"PE", "Production Engineering"}, {"ITEP", "4-year Integrated Teacher Education Programme (ITEP) B.Sc-B.Ed"},
	}
	branchIDs := make(map[string]uuid.UUID)
	for _, u := range units {
		var id uuid.UUID
		if err := db.Pool.QueryRow(ctx, `INSERT INTO branches(code,name,active) VALUES($1,$2,TRUE) ON CONFLICT(code) DO UPDATE SET name=EXCLUDED.name RETURNING id`, u.Code, u.Name).Scan(&id); err != nil { return fmt.Errorf("seed branch %s: %w", u.Code, err) }
		branchIDs[u.Code] = id
		unitCode := u.Code; if unitCode == "CH" { unitCode = "CHE" }
		typeName := "department"; if u.Code == "BARCH" || u.Code == "ITEP" { typeName = "academic_programme" }
		if _, err := db.Pool.Exec(ctx, `INSERT INTO academic_units(code,name,type,active) VALUES($1,$2,$3,TRUE) ON CONFLICT(code) DO UPDATE SET name=EXCLUDED.name,type=EXCLUDED.type,active=TRUE`, unitCode,u.Name,typeName); err != nil { return fmt.Errorf("seed academic unit %s: %w", unitCode, err) }
	}

	// These are demo entries only. Production must use verified NITC curriculum data.
	courses := []courseSeed{
		{"CS201", "Data Structures", "Demo catalog entry.", []string{"CSE"}},
		{"EC301", "Signals & Systems", "Demo catalog entry.", []string{"ECE", "EEE"}},
		{"ME203", "Thermodynamics", "Demo catalog entry.", []string{"ME", "CH"}},
		{"MA101", "Mathematics I", "Demo catalog entry.", []string{"CSE", "ECE", "EEE", "ME", "CE", "CH", "BT", "BARCH", "ENE", "EP", "MSE", "PE"}},
	}
	for _, c := range courses {
		var id uuid.UUID
		if err := db.Pool.QueryRow(ctx, `INSERT INTO courses(code,name,description,active) VALUES($1,$2,$3,TRUE) ON CONFLICT(code) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description RETURNING id`, c.Code,c.Name,c.Description).Scan(&id); err != nil { return fmt.Errorf("seed course %s: %w", c.Code, err) }
		for _, code := range c.Branches {
			if bid, ok := branchIDs[code]; ok { if _, err := db.Pool.Exec(ctx, `INSERT INTO course_branches(course_id,branch_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, bid); err != nil { return fmt.Errorf("map course %s to branch %s: %w", c.Code, code, err) } }
			unitCode := code; if unitCode == "CH" { unitCode = "CHE" }
			if _, err := db.Pool.Exec(ctx, `INSERT INTO course_academic_units(course_id,academic_unit_id) SELECT $1,id FROM academic_units WHERE code=$2 ON CONFLICT DO NOTHING`, id, unitCode); err != nil { return fmt.Errorf("map course %s to academic unit %s: %w", c.Code, unitCode, err) }
		}
	}

	if os.Getenv("APP_ENV") != "production" {
		fixturePath := filepath.Join("fixtures", "sample.pdf")
		if f, err := os.Open(fixturePath); err == nil {
			if st, statErr := f.Stat(); statErr == nil { _ = store.Put(ctx, "sample.pdf", f, st.Size(), "application/pdf") }
			_ = f.Close()
		}
	}
	return nil
}
