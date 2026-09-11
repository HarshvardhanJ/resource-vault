package catalog

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBranches(ctx context.Context) ([]Branch, error) {
	return s.repo.ListActiveBranches(ctx)
}

func (s *Service) GetBranch(ctx context.Context, code string) (*Branch, error) {
	return s.repo.GetBranchByCode(ctx, code)
}

func (s *Service) GetCourse(ctx context.Context, code string) (*Course, error) {
	return s.repo.GetCourseByCode(ctx, code)
}

func (s *Service) ListCoursesByBranch(ctx context.Context, branchCode string) ([]Course, error) {
	return s.repo.ListCoursesByBranch(ctx, branchCode)
}

func (s *Service) SearchCourses(ctx context.Context, query string, limit int) ([]Course, error) {
	return s.repo.SearchCourses(ctx, query, limit)
}
