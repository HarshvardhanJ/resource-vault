package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AcademicUnitRepository struct { pool *pgxpool.Pool }
func NewAcademicUnitRepository(pool *pgxpool.Pool) *AcademicUnitRepository { return &AcademicUnitRepository{pool:pool} }

func (r *AcademicUnitRepository) ListActive(ctx context.Context) ([]AcademicUnit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT au.id, au.code, au.name, au.type, au.active,
		       COUNT(DISTINCT c.id) FILTER (WHERE c.active=TRUE) AS course_count
		FROM academic_units au
		LEFT JOIN course_academic_units cau ON cau.academic_unit_id=au.id
		LEFT JOIN courses c ON c.id=cau.course_id
		WHERE au.active=TRUE
		GROUP BY au.id, au.code, au.name, au.type, au.active
		ORDER BY CASE au.type WHEN 'academic_programme' THEN 0 ELSE 1 END, au.name ASC
	`)
	if err != nil { return nil, fmt.Errorf("list academic units: %w",err) }
	defer rows.Close()
	var units []AcademicUnit
	for rows.Next(){
		var u AcademicUnit
		if err:=rows.Scan(&u.ID,&u.Code,&u.Name,&u.Type,&u.Active,&u.CourseCount);err!=nil{return nil,fmt.Errorf("scan academic unit: %w",err)}
		units=append(units,u)
	}
	return units,rows.Err()
}

func (r *AcademicUnitRepository) GetByCode(ctx context.Context, code string) (*AcademicUnit,error){
	var u AcademicUnit
	err:=r.pool.QueryRow(ctx,`SELECT id,code,name,type,active FROM academic_units WHERE UPPER(code)=$1 AND active=TRUE`,strings.ToUpper(strings.TrimSpace(code))).Scan(&u.ID,&u.Code,&u.Name,&u.Type,&u.Active)
	if err!=nil{return nil,err};return &u,nil
}
