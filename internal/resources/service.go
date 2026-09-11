package resources

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID, onlyPublished bool) (*Resource, error) {
	return s.repo.GetByID(ctx, id, onlyPublished)
}

func (s *Service) ListByCourse(ctx context.Context, courseID uuid.UUID, examType string, year int, semester string) ([]Resource, error) {
	return s.repo.ListPublishedByCourse(ctx, courseID, examType, year, semester)
}

func (s *Service) ListRecent(ctx context.Context, limit int) ([]Resource, error) {
	return s.repo.ListRecentPublished(ctx, limit)
}

func (s *Service) Search(ctx context.Context, params FilterParams) ([]Resource, int, error) {
	return s.repo.Search(ctx, params)
}

func (s *Service) CheckDuplicate(ctx context.Context, sha256 string) (bool, *Resource, error) {
	return s.repo.CheckDuplicate(ctx, sha256)
}
