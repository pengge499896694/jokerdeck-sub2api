package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type FinanceFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
	UserID    int64
	GroupID   int64
	Search    string
	Page      int
	PageSize  int
}

type FinanceUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	GroupID   *int64    `json:"group_id,omitempty"`
	GroupName string    `json:"group_name,omitempty"`
}

type FinanceGroup struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	SortOrder   int           `json:"sort_order"`
	Members     []FinanceUser `json:"members"`
}

type FinanceRecharge struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	GroupName string     `json:"group_name"`
	Type      string     `json:"type"`
	Value     float64    `json:"value"`
	Status    string     `json:"status"`
	Notes     string     `json:"notes"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type FinanceUsage struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	GroupName    string    `json:"group_name"`
	Model        string    `json:"model"`
	RequestID    string    `json:"request_id"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	ActualCost   float64   `json:"actual_cost"`
	CreatedAt    time.Time `json:"created_at"`
}

type FinanceRepository interface {
	ListUsers(ctx context.Context, search string) ([]FinanceUser, error)
	ListGroups(ctx context.Context) ([]FinanceGroup, error)
	CreateGroup(ctx context.Context, actorID int64, name, description string) (*FinanceGroup, error)
	UpdateGroup(ctx context.Context, id int64, name, description string) (*FinanceGroup, error)
	DeleteGroup(ctx context.Context, id int64) error
	ReplaceGroupMembers(ctx context.Context, groupID int64, userIDs []int64) error
	ListRecharges(ctx context.Context, filters FinanceFilters) ([]FinanceRecharge, *pagination.PaginationResult, float64, error)
	ListUsage(ctx context.Context, filters FinanceFilters) ([]FinanceUsage, *pagination.PaginationResult, float64, error)
}

type FinanceService struct {
	repo FinanceRepository
}

func NewFinanceService(repo FinanceRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

func (s *FinanceService) ListUsers(ctx context.Context, search string) ([]FinanceUser, error) {
	return s.repo.ListUsers(ctx, search)
}

func (s *FinanceService) ListGroups(ctx context.Context) ([]FinanceGroup, error) {
	return s.repo.ListGroups(ctx)
}

func (s *FinanceService) CreateGroup(ctx context.Context, actorID int64, name, description string) (*FinanceGroup, error) {
	return s.repo.CreateGroup(ctx, actorID, name, description)
}

func (s *FinanceService) UpdateGroup(ctx context.Context, id int64, name, description string) (*FinanceGroup, error) {
	return s.repo.UpdateGroup(ctx, id, name, description)
}

func (s *FinanceService) DeleteGroup(ctx context.Context, id int64) error {
	return s.repo.DeleteGroup(ctx, id)
}

func (s *FinanceService) ReplaceGroupMembers(ctx context.Context, groupID int64, userIDs []int64) error {
	return s.repo.ReplaceGroupMembers(ctx, groupID, userIDs)
}

func (s *FinanceService) ListRecharges(ctx context.Context, filters FinanceFilters) ([]FinanceRecharge, *pagination.PaginationResult, float64, error) {
	return s.repo.ListRecharges(ctx, filters)
}

func (s *FinanceService) ListUsage(ctx context.Context, filters FinanceFilters) ([]FinanceUsage, *pagination.PaginationResult, float64, error) {
	return s.repo.ListUsage(ctx, filters)
}
