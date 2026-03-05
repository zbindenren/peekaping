package incident

import (
	"context"

	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, dto *CreateDto) (*Model, error)
	FindByID(ctx context.Context, id string) (*Model, error)
	FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error)
	FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error)
	FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error)
	FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error)
	Update(ctx context.Context, id string, dto *UpdateDto) (*Model, error)
	Resolve(ctx context.Context, id string) (*Model, error)
	Delete(ctx context.Context, id string) error
	DeleteByStatusPageID(ctx context.Context, statusPageID string) error
}

type ServiceImpl struct {
	repository Repository
	logger     *zap.SugaredLogger
}

func NewService(repository Repository, logger *zap.SugaredLogger) Service {
	return &ServiceImpl{
		repository: repository,
		logger:     logger.Named("[incident-service]"),
	}
}

func (s *ServiceImpl) Create(ctx context.Context, dto *CreateDto) (*Model, error) {
	return s.repository.Create(ctx, dto)
}

func (s *ServiceImpl) FindByID(ctx context.Context, id string) (*Model, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *ServiceImpl) FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error) {
	return s.repository.FindAll(ctx, page, limit, q)
}

func (s *ServiceImpl) FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error) {
	return s.repository.FindByStatusPageID(ctx, statusPageID, page, limit)
}

func (s *ServiceImpl) FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error) {
	return s.repository.FindActiveByStatusPageID(ctx, statusPageID)
}

func (s *ServiceImpl) FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error) {
	return s.repository.FindByStatusPageSlug(ctx, slug, page, limit)
}

func (s *ServiceImpl) Update(ctx context.Context, id string, dto *UpdateDto) (*Model, error) {
	return s.repository.Update(ctx, id, dto)
}

func (s *ServiceImpl) Resolve(ctx context.Context, id string) (*Model, error) {
	return s.repository.Resolve(ctx, id)
}

func (s *ServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func (s *ServiceImpl) DeleteByStatusPageID(ctx context.Context, statusPageID string) error {
	return s.repository.DeleteByStatusPageID(ctx, statusPageID)
}
