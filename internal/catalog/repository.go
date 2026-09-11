package catalog

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("catalog entity not found")

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) ListActiveBranches(ctx context.Context) ([]Branch, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT b.id, b.code, b.name, b.active, b.created_at,
               COUNT(cb.course_id) AS course_count
        FROM branches b
        LEFT JOIN course_branches cb ON b.id = cb.branch_id
        WHERE b.active = TRUE
        GROUP BY b.id, b.code, b.name, b.active, b.created_at
        ORDER BY b.code ASC
    `)
    if err != nil { return nil, fmt.Errorf("failed to query branches: %w", err) }
    defer rows.Close()
    var branches []Branch
    for rows.Next() {
        var b Branch
        if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Active, &b.CreatedAt, &b.CourseCount); err != nil {
            return nil, fmt.Errorf("failed to scan branch: %w", err)
        }
        branches = append(branches, b)
    }
    return branches, rows.Err()
}

func (r *Repository) GetBranchByCode(ctx context.Context, code string) (*Branch, error) {
    upperCode := strings.ToUpper(strings.TrimSpace(code))
    var b Branch
    err := r.pool.QueryRow(ctx, `
        SELECT b.id, b.code, b.name, b.active, b.created_at,
               COUNT(cb.course_id) AS course_count
        FROM branches b
        LEFT JOIN course_branches cb ON b.id = cb.branch_id
        WHERE UPPER(b.code) = $1 AND b.active = TRUE
        GROUP BY b.id, b.code, b.name, b.active, b.created_at
    `, upperCode).Scan(&b.ID, &b.Code, &b.Name, &b.Active, &b.CreatedAt, &b.CourseCount)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) { return nil, ErrNotFound }
        return nil, fmt.Errorf("failed to query branch by code: %w", err)
    }
    return &b, nil
}

func (r *Repository) ListCoursesByBranch(ctx context.Context, branchCode string) ([]Course, error) {
    upperCode := strings.ToUpper(strings.TrimSpace(branchCode))
    rows, err := r.pool.Query(ctx, `
        SELECT c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at,
               COUNT(DISTINCT res.id) FILTER (WHERE res.status = 'PUBLISHED') AS resource_count
        FROM courses c
        JOIN course_branches cb ON c.id = cb.course_id
        JOIN branches b ON cb.branch_id = b.id
        LEFT JOIN resources res ON c.id = res.course_id
        WHERE UPPER(b.code) = $1 AND c.active = TRUE
        GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
        ORDER BY c.code ASC
    `, upperCode)
    if err != nil { return nil, fmt.Errorf("failed to query courses for branch: %w", err) }
    defer rows.Close()
    var courses []Course
    for rows.Next() {
        var c Course
        if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Active, &c.CreatedAt, &c.UpdatedAt, &c.ResourceCount); err != nil {
            return nil, fmt.Errorf("failed to scan course: %w", err)
        }
        courses = append(courses, c)
    }
    return courses, rows.Err()
}

func (r *Repository) GetCourseByCode(ctx context.Context, code string) (*Course, error) {
    upperCode := strings.ToUpper(strings.TrimSpace(code))
    var c Course
    err := r.pool.QueryRow(ctx, `
        SELECT c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at,
               COUNT(DISTINCT res.id) FILTER (WHERE res.status = 'PUBLISHED') AS resource_count
        FROM courses c
        LEFT JOIN resources res ON c.id = res.course_id
        WHERE UPPER(c.code) = $1 AND c.active = TRUE
        GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
    `, upperCode).Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Active, &c.CreatedAt, &c.UpdatedAt, &c.ResourceCount)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) { return nil, ErrNotFound }
        return nil, fmt.Errorf("failed to query course by code: %w", err)
    }

    bRows, err := r.pool.Query(ctx, `
        SELECT b.id, b.code, b.name, b.active, b.created_at
        FROM branches b
        JOIN course_branches cb ON b.id = cb.branch_id
        WHERE cb.course_id = $1 AND b.active = TRUE
        ORDER BY b.code ASC
    `, c.ID)
    if err == nil {
        defer bRows.Close()
        for bRows.Next() {
            var b Branch
            if err := bRows.Scan(&b.ID, &b.Code, &b.Name, &b.Active, &b.CreatedAt); err == nil { c.Branches = append(c.Branches, b) }
        }
    }

    aRows, err := r.pool.Query(ctx, `
        SELECT au.id, au.code, au.name, au.type, au.active, COUNT(DISTINCT c2.id)
        FROM academic_units au
        JOIN course_academic_units cau ON au.id = cau.academic_unit_id
        LEFT JOIN course_academic_units cau2 ON cau2.academic_unit_id = au.id
        LEFT JOIN courses c2 ON c2.id = cau2.course_id AND c2.active = TRUE
        WHERE cau.course_id = $1 AND au.active = TRUE
        GROUP BY au.id, au.code, au.name, au.type, au.active
        ORDER BY au.name ASC
    `, c.ID)
    if err == nil {
        defer aRows.Close()
        for aRows.Next() {
            var u AcademicUnit
            if err := aRows.Scan(&u.ID, &u.Code, &u.Name, &u.Type, &u.Active, &u.CourseCount); err == nil { c.AcademicUnits = append(c.AcademicUnits, u) }
        }
    }

    return &c, nil
}

func (r *Repository) SearchCourses(ctx context.Context, query string, limit int) ([]Course, error) {
    if limit <= 0 || limit > 50 { limit = 20 }
    q := strings.TrimSpace(query)
    searchPattern := "%" + q + "%"
    prefixPattern := strings.ToUpper(q) + "%"
    rows, err := r.pool.Query(ctx, `
        SELECT c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at,
               COUNT(DISTINCT res.id) FILTER (WHERE res.status = 'PUBLISHED') AS resource_count
        FROM courses c
        LEFT JOIN resources res ON c.id = res.course_id
        WHERE c.active = TRUE
          AND (UPPER(c.code) LIKE $1 OR c.name ILIKE $2 OR c.code ILIKE $2)
        GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
        ORDER BY CASE WHEN UPPER(c.code) = UPPER($3) THEN 1 WHEN UPPER(c.code) LIKE $1 THEN 2 ELSE 3 END, c.code ASC
        LIMIT $4
    `, prefixPattern, searchPattern, q, limit)
    if err != nil { return nil, fmt.Errorf("failed to search courses: %w", err) }
    defer rows.Close()
    var courses []Course
    for rows.Next() {
        var c Course
        if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Active, &c.CreatedAt, &c.UpdatedAt, &c.ResourceCount); err != nil {
            return nil, fmt.Errorf("failed to scan course: %w", err)
        }
        courses = append(courses, c)
    }
    return courses, rows.Err()
}

func (r *Repository) UpdateCourse(ctx context.Context, oldCode, newCode, name, description string, branchCodes []string) (*Course, error) {
    upperOld := strings.ToUpper(strings.TrimSpace(oldCode))
    upperNew := strings.ToUpper(strings.TrimSpace(newCode))
    trimmedName := strings.TrimSpace(name)
    trimmedDesc := strings.TrimSpace(description)
    tx, err := r.pool.Begin(ctx)
    if err != nil { return nil, fmt.Errorf("failed to begin transaction: %w", err) }
    defer tx.Rollback(ctx)

    var courseID uuid.UUID
    err = tx.QueryRow(ctx, `UPDATE courses SET code=$1,name=$2,description=$3,updated_at=NOW() WHERE UPPER(code)=$4 AND active=TRUE RETURNING id`, upperNew, trimmedName, trimmedDesc, upperOld).Scan(&courseID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) { return nil, ErrNotFound }
        return nil, fmt.Errorf("failed to update course: %w", err)
    }
    _, err = tx.Exec(ctx, "DELETE FROM course_branches WHERE course_id=$1", courseID)
    if err != nil { return nil, fmt.Errorf("failed to clear old course branches: %w", err) }
    for _, bCode := range branchCodes {
        clean := strings.ToUpper(strings.TrimSpace(bCode)); if clean == "" { continue }
        if _, err = tx.Exec(ctx, `INSERT INTO course_branches(course_id,branch_id) SELECT $1,id FROM branches WHERE UPPER(code)=$2 ON CONFLICT DO NOTHING`, courseID, clean); err != nil {
            return nil, fmt.Errorf("failed to map branch %s: %w", bCode, err)
        }
    }
    meta, _ := json.Marshal(map[string]interface{}{"old_code":upperOld,"new_code":upperNew,"name":trimmedName,"branches":branchCodes})
    _, _ = tx.Exec(ctx, `INSERT INTO audit_events(action,entity_type,entity_id,metadata_json,created_at) VALUES('COURSE_UPDATED','course',$1,$2,NOW())`, courseID, meta)
    if err := tx.Commit(ctx); err != nil { return nil, fmt.Errorf("failed to commit course update: %w", err) }
    return r.GetCourseByCode(ctx, upperNew)
}

func (r *Repository) CreateCourse(ctx context.Context, code, name, description string, branchCodes []string) (*Course, error) {
    upperCode := strings.ToUpper(strings.TrimSpace(code)); trimmedName := strings.TrimSpace(name); trimmedDesc := strings.TrimSpace(description)
    tx, err := r.pool.Begin(ctx); if err != nil { return nil, fmt.Errorf("failed to begin transaction: %w", err) }; defer tx.Rollback(ctx)
    var courseID uuid.UUID
    if err = tx.QueryRow(ctx, `INSERT INTO courses(code,name,description,active) VALUES($1,$2,$3,TRUE) RETURNING id`, upperCode, trimmedName, trimmedDesc).Scan(&courseID); err != nil { return nil, fmt.Errorf("failed to insert course: %w", err) }
    for _, bCode := range branchCodes { clean := strings.ToUpper(strings.TrimSpace(bCode)); if clean == "" { continue }; if _, err = tx.Exec(ctx, `INSERT INTO course_branches(course_id,branch_id) SELECT $1,id FROM branches WHERE UPPER(code)=$2 ON CONFLICT DO NOTHING`, courseID, clean); err != nil { return nil, fmt.Errorf("failed to map branch %s: %w", bCode, err) } }
    meta, _ := json.Marshal(map[string]interface{}{"code":upperCode,"name":trimmedName,"branches":branchCodes})
    _, _ = tx.Exec(ctx, `INSERT INTO audit_events(action,entity_type,entity_id,metadata_json,created_at) VALUES('COURSE_CREATED','course',$1,$2,NOW())`, courseID, meta)
    if err := tx.Commit(ctx); err != nil { return nil, fmt.Errorf("failed to commit course creation: %w", err) }
    return r.GetCourseByCode(ctx, upperCode)
}
