package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("catalog entity not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

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
	if err != nil {
		return nil, fmt.Errorf("failed to query branches: %w", err)
	}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query branch by code: %w", err)
	}
	return &b, nil
}

func (r *Repository) ListCoursesByBranch(ctx context.Context, branchCode string) ([]Course, error) {
	upperCode := strings.ToUpper(strings.TrimSpace(branchCode))
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at,
		       COUNT(DISTINCT r.id) FILTER (WHERE r.status = 'PUBLISHED') AS resource_count
		FROM courses c
		JOIN course_branches cb ON c.id = cb.course_id
		JOIN branches b ON cb.branch_id = b.id
		LEFT JOIN resources r ON c.id = r.course_id
		WHERE UPPER(b.code) = $1 AND c.active = TRUE
		GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
		ORDER BY c.code ASC
	`, upperCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query courses for branch: %w", err)
	}
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
		       COUNT(DISTINCT r.id) FILTER (WHERE r.status = 'PUBLISHED') AS resource_count
		FROM courses c
		LEFT JOIN resources r ON c.id = r.course_id
		WHERE UPPER(c.code) = $1 AND c.active = TRUE
		GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
	`, upperCode).Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Active, &c.CreatedAt, &c.UpdatedAt, &c.ResourceCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query course by code: %w", err)
	}

	// Fetch mapped branches
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
			if err := bRows.Scan(&b.ID, &b.Code, &b.Name, &b.Active, &b.CreatedAt); err == nil {
				c.Branches = append(c.Branches, b)
			}
		}
	}

	return &c, nil
}

func (r *Repository) SearchCourses(ctx context.Context, query string, limit int) ([]Course, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := strings.TrimSpace(query)
	searchPattern := "%" + q + "%"
	prefixPattern := strings.ToUpper(q) + "%"

	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at,
		       COUNT(DISTINCT r.id) FILTER (WHERE r.status = 'PUBLISHED') AS resource_count
		FROM courses c
		LEFT JOIN resources r ON c.id = r.course_id
		WHERE c.active = TRUE
		  AND (
		      UPPER(c.code) LIKE $1
		      OR c.name ILIKE $2
		      OR c.code ILIKE $2
		  )
		GROUP BY c.id, c.code, c.name, c.description, c.active, c.created_at, c.updated_at
		ORDER BY 
		    CASE WHEN UPPER(c.code) = UPPER($3) THEN 1
		         WHEN UPPER(c.code) LIKE $1 THEN 2
		         ELSE 3 END,
		    c.code ASC
		LIMIT $4
	`, prefixPattern, searchPattern, q, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search courses: %w", err)
	}
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
