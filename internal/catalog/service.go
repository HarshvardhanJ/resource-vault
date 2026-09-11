package catalog

import (
	"context"
	"errors"
	"strings"
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

func (s *Service) CreateCourse(ctx context.Context, code, name, description string, branchCodes []string) (*Course, error) {
	cleanCode := strings.ToUpper(strings.TrimSpace(code))
	cleanName := strings.TrimSpace(name)
	if cleanCode == "" {
		return nil, errors.New("course code cannot be empty")
	}
	if cleanName == "" {
		return nil, errors.New("course name cannot be empty")
	}
	return s.repo.CreateCourse(ctx, cleanCode, cleanName, description, branchCodes)
}

func (s *Service) UpdateCourse(ctx context.Context, oldCode, newCode, name, description string, branchCodes []string) (*Course, error) {
	cleanNewCode := strings.ToUpper(strings.TrimSpace(newCode))
	cleanName := strings.TrimSpace(name)
	if cleanNewCode == "" {
		return nil, errors.New("course code cannot be empty")
	}
	if cleanName == "" {
		return nil, errors.New("course name cannot be empty")
	}
	return s.repo.UpdateCourse(ctx, oldCode, cleanNewCode, cleanName, description, branchCodes)
}
