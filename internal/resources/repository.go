package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("resource not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, onlyPublished bool) (*Resource, error) {
	query := `
		SELECT r.id, r.course_id, r.uploader_id, r.resource_type, r.semester,
		       r.academic_year_start, r.exam_type, r.title, r.file_size_bytes, r.sha256,
		       r.storage_provider, r.storage_identifier, r.storage_filename,
		       r.status, r.archive_status, r.scan_status,
		       r.created_at, r.updated_at, r.published_at, r.removed_at,
		       c.code AS course_code, c.name AS course_name
		FROM resources r
		JOIN courses c ON r.course_id = c.id
		WHERE r.id = $1
	`
	if onlyPublished {
		query += " AND r.status = 'PUBLISHED'"
	}

	var res Resource
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.CourseID, &res.UploaderID, &res.ResourceType, &res.Semester,
		&res.AcademicYearStart, &res.ExamType, &res.Title, &res.FileSizeBytes, &res.SHA256,
		&res.StorageProvider, &res.StorageIdentifier, &res.StorageFilename,
		&res.Status, &res.ArchiveStatus, &res.ScanStatus,
		&res.CreatedAt, &res.UpdatedAt, &res.PublishedAt, &res.RemovedAt,
		&res.CourseCode, &res.CourseName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query resource: %w", err)
	}

	// Fetch branch codes for this course
	bRows, err := r.pool.Query(ctx, `
		SELECT b.code
		FROM branches b
		JOIN course_branches cb ON b.id = cb.branch_id
		WHERE cb.course_id = $1 AND b.active = TRUE
		ORDER BY b.code ASC
	`, res.CourseID)
	if err == nil {
		defer bRows.Close()
		for bRows.Next() {
			var bCode string
			if err := bRows.Scan(&bCode); err == nil {
				res.BranchCodes = append(res.BranchCodes, bCode)
			}
		}
	}

	return &res, nil
}

func (r *Repository) ListPublishedByCourse(ctx context.Context, courseID uuid.UUID, examType string, year int, semester string) ([]Resource, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("r.course_id = $%d", argIdx))
	args = append(args, courseID)
	argIdx++

	conditions = append(conditions, "r.status = 'PUBLISHED'")

	if examType != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.exam_type) = $%d", argIdx))
		args = append(args, strings.ToUpper(examType))
		argIdx++
	}

	if year > 0 {
		conditions = append(conditions, fmt.Sprintf("r.academic_year_start = $%d", argIdx))
		args = append(args, year)
		argIdx++
	}

	if semester != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.semester) = $%d", argIdx))
		args = append(args, strings.ToUpper(semester))
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.course_id, r.uploader_id, r.resource_type, r.semester,
		       r.academic_year_start, r.exam_type, r.title, r.file_size_bytes, r.sha256,
		       r.storage_provider, r.storage_identifier, r.storage_filename,
		       r.status, r.archive_status, r.scan_status,
		       r.created_at, r.updated_at, r.published_at, r.removed_at,
		       c.code AS course_code, c.name AS course_name
		FROM resources r
		JOIN courses c ON r.course_id = c.id
		WHERE %s
		ORDER BY r.academic_year_start DESC, 
		         CASE r.exam_type WHEN 'ENDSEM' THEN 1 WHEN 'MIDSEM' THEN 2 WHEN 'QUIZ' THEN 3 ELSE 4 END,
		         r.created_at DESC
	`, strings.Join(conditions, " AND "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list published resources for course: %w", err)
	}
	defer rows.Close()

	var results []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.CourseID, &res.UploaderID, &res.ResourceType, &res.Semester,
			&res.AcademicYearStart, &res.ExamType, &res.Title, &res.FileSizeBytes, &res.SHA256,
			&res.StorageProvider, &res.StorageIdentifier, &res.StorageFilename,
			&res.Status, &res.ArchiveStatus, &res.ScanStatus,
			&res.CreatedAt, &res.UpdatedAt, &res.PublishedAt, &res.RemovedAt,
			&res.CourseCode, &res.CourseName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan resource: %w", err)
		}
		results = append(results, res)
	}

	return results, rows.Err()
}

func (r *Repository) ListRecentPublished(ctx context.Context, limit int) ([]Resource, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.course_id, r.uploader_id, r.resource_type, r.semester,
		       r.academic_year_start, r.exam_type, r.title, r.file_size_bytes, r.sha256,
		       r.storage_provider, r.storage_identifier, r.storage_filename,
		       r.status, r.archive_status, r.scan_status,
		       r.created_at, r.updated_at, r.published_at, r.removed_at,
		       c.code AS course_code, c.name AS course_name
		FROM resources r
		JOIN courses c ON r.course_id = c.id
		WHERE r.status = 'PUBLISHED'
		ORDER BY r.published_at DESC NULLS LAST, r.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent published resources: %w", err)
	}
	defer rows.Close()

	var results []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.CourseID, &res.UploaderID, &res.ResourceType, &res.Semester,
			&res.AcademicYearStart, &res.ExamType, &res.Title, &res.FileSizeBytes, &res.SHA256,
			&res.StorageProvider, &res.StorageIdentifier, &res.StorageFilename,
			&res.Status, &res.ArchiveStatus, &res.ScanStatus,
			&res.CreatedAt, &res.UpdatedAt, &res.PublishedAt, &res.RemovedAt,
			&res.CourseCode, &res.CourseName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recent resource: %w", err)
		}
		results = append(results, res)
	}

	return results, rows.Err()
}

func (r *Repository) Search(ctx context.Context, params FilterParams) ([]Resource, int, error) {
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "r.status = 'PUBLISHED'")

	if params.Query != "" {
		q := strings.TrimSpace(params.Query)
		pattern := "%" + q + "%"
		prefix := strings.ToUpper(q) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			UPPER(c.code) LIKE $%d 
			OR c.name ILIKE $%d 
			OR r.title ILIKE $%d 
			OR CAST(r.academic_year_start AS TEXT) LIKE $%d
		)`, argIdx, argIdx+1, argIdx+1, argIdx))
		args = append(args, prefix, pattern)
		argIdx += 2
	}

	if params.CourseCode != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(c.code) = $%d", argIdx))
		args = append(args, strings.ToUpper(params.CourseCode))
		argIdx++
	}

	if params.ExamType != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.exam_type) = $%d", argIdx))
		args = append(args, strings.ToUpper(params.ExamType))
		argIdx++
	}

	if params.Semester != "" {
		conditions = append(conditions, fmt.Sprintf("UPPER(r.semester) = $%d", argIdx))
		args = append(args, strings.ToUpper(params.Semester))
		argIdx++
	}

	if params.Year > 0 {
		conditions = append(conditions, fmt.Sprintf("r.academic_year_start = $%d", argIdx))
		args = append(args, params.Year)
		argIdx++
	}

	var joinBranch string
	if params.BranchCode != "" {
		joinBranch = `
			JOIN course_branches cb ON c.id = cb.course_id
			JOIN branches b ON cb.branch_id = b.id AND UPPER(b.code) = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, strings.ToUpper(params.BranchCode))
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT r.id)
		FROM resources r
		JOIN courses c ON r.course_id = c.id
		%s
		WHERE %s
	`, joinBranch, whereClause)

	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT r.id, r.course_id, r.uploader_id, r.resource_type, r.semester,
		       r.academic_year_start, r.exam_type, r.title, r.file_size_bytes, r.sha256,
		       r.storage_provider, r.storage_identifier, r.storage_filename,
		       r.status, r.archive_status, r.scan_status,
		       r.created_at, r.updated_at, r.published_at, r.removed_at,
		       c.code AS course_code, c.name AS course_name
		FROM resources r
		JOIN courses c ON r.course_id = c.id
		%s
		WHERE %s
		ORDER BY r.academic_year_start DESC, r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, joinBranch, whereClause, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	var results []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.CourseID, &res.UploaderID, &res.ResourceType, &res.Semester,
			&res.AcademicYearStart, &res.ExamType, &res.Title, &res.FileSizeBytes, &res.SHA256,
			&res.StorageProvider, &res.StorageIdentifier, &res.StorageFilename,
			&res.Status, &res.ArchiveStatus, &res.ScanStatus,
			&res.CreatedAt, &res.UpdatedAt, &res.PublishedAt, &res.RemovedAt,
			&res.CourseCode, &res.CourseName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, res)
	}

	return results, totalCount, rows.Err()
}

func (r *Repository) CheckDuplicate(ctx context.Context, sha256 string) (bool, *Resource, error) {
	var res Resource
	err := r.pool.QueryRow(ctx, `
		SELECT r.id, r.course_id, r.status, r.title, r.sha256
		FROM resources r
		WHERE r.sha256 = $1 AND r.status != 'REMOVED'
		LIMIT 1
	`, sha256).Scan(&res.ID, &res.CourseID, &res.Status, &res.Title, &res.SHA256)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &res, nil
}
