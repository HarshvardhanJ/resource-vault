package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AcademicUnitRepository struct { pool *pgxpool.Pool }
func NewAcademicUnitRepository(pool *pgxpool.Pool) *AcademicUnitRepository { return &AcademicUnitRepository{pool:pool} }

func (r *AcademicUnitRepository) ListActive(ctx context.Context) ([]AcademicUnit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, type, active FROM academic_units WHERE active=TRUE ORDER BY CASE type WHEN 'academic_programme' THEN 0 ELSE 1 END, name ASC`)
	if err != nil { return nil, fmt.Errorf("list academic units: %w",err) }
	defer rows.Close()
	var units []AcademicUnit
	for rows.Next(){ var u AcademicUnit; if err:=rows.Scan(&u.ID,&u.Code,&u.Name,&u.Type,&u.Active);err!=nil{return nil,fmt.Errorf("scan academic unit: %w",err)};units=append(units,u) }
	return units,rows.Err()
}

func (r *AcademicUnitRepository) GetByCode(ctx context.Context, code string) (*AcademicUnit,error){
	var u AcademicUnit
	err:=r.pool.QueryRow(ctx,`SELECT id,code,name,type,active FROM academic_units WHERE UPPER(code)=$1 AND active=TRUE`,strings.ToUpper(strings.TrimSpace(code))).Scan(&u.ID,&u.Code,&u.Name,&u.Type,&u.Active)
	if err!=nil{return nil,err};return &u,nil
}
