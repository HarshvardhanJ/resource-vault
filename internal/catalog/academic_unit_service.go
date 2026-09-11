package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type AcademicUnitService struct {
	repo *AcademicUnitRepository
}

func NewAcademicUnitService(repo *AcademicUnitRepository) *AcademicUnitService {
	return &AcademicUnitService{repo: repo}
}

func (s *AcademicUnitService) ListActive(ctx context.Context) ([]AcademicUnit, error) {
	return s.repo.ListActive(ctx)
}

func (s *AcademicUnitService) ValidateCode(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" { return errors.New("academic unit code cannot be empty") }
	if _, err := s.repo.GetByCode(ctx, code); err != nil { return fmt.Errorf("unknown academic unit: %w", err) }
	return nil
}
