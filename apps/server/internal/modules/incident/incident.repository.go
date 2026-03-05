package incident

import "context"

type Repository interface {
	Create(ctx context.Context, entity *CreateDto) (*Model, error)
	FindByID(ctx context.Context, id string) (*Model, error)
	FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error)
	FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error)
	FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error)
	FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error)
	Update(ctx context.Context, id string, entity *UpdateDto) (*Model, error)
	Resolve(ctx context.Context, id string) (*Model, error)
	Delete(ctx context.Context, id string) error
	DeleteByStatusPageID(ctx context.Context, statusPageID string) error
}
