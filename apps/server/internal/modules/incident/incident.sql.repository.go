package incident

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var _ Repository = (*SQLRepositoryImpl)(nil)

type sqlModel struct {
	bun.BaseModel `bun:"table:incidents,alias:i"`

	ID           string     `bun:"id,pk"`
	StatusPageID string     `bun:"status_page_id,notnull"`
	Title        string     `bun:"title,notnull"`
	Content      string     `bun:"content,notnull,default:''"`
	Style        string     `bun:"style,notnull,default:'warning'"`
	Active       bool       `bun:"active,notnull,default:true"`
	ResolvedAt   *time.Time `bun:"resolved_at"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

func toDomainModelFromSQL(sm *sqlModel) *Model {
	return &Model{
		ID:           sm.ID,
		StatusPageID: sm.StatusPageID,
		Title:        sm.Title,
		Content:      sm.Content,
		Style:        sm.Style,
		Active:       sm.Active,
		ResolvedAt:   sm.ResolvedAt,
		CreatedAt:    sm.CreatedAt,
		UpdatedAt:    sm.UpdatedAt,
	}
}

type SQLRepositoryImpl struct {
	db *bun.DB
}

func NewSQLRepository(db *bun.DB) Repository {
	return &SQLRepositoryImpl{db: db}
}

func (r *SQLRepositoryImpl) Create(ctx context.Context, entity *CreateDto) (*Model, error) {
	sm := &sqlModel{
		ID:           uuid.New().String(),
		StatusPageID: entity.StatusPageID,
		Title:        entity.Title,
		Content:      entity.Content,
		Style:        entity.Style,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.NewInsert().Model(sm).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}

	return toDomainModelFromSQL(sm), nil
}

func (r *SQLRepositoryImpl) FindByID(ctx context.Context, id string) (*Model, error) {
	sm := new(sqlModel)
	err := r.db.NewSelect().Model(sm).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return toDomainModelFromSQL(sm), nil
}

func (r *SQLRepositoryImpl) FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error) {
	query := r.db.NewSelect().Model((*sqlModel)(nil))

	if q != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(q)+"%")
	}

	query = query.Order("created_at DESC").
		Limit(limit).
		Offset(page * limit)

	var sms []*sqlModel
	err := query.Scan(ctx, &sms)
	if err != nil {
		return nil, err
	}

	var models []*Model
	for _, sm := range sms {
		models = append(models, toDomainModelFromSQL(sm))
	}
	return models, nil
}

func (r *SQLRepositoryImpl) FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error) {
	var sms []*sqlModel
	err := r.db.NewSelect().
		Model(&sms).
		Where("status_page_id = ?", statusPageID).
		Order("created_at DESC").
		Limit(limit).
		Offset(page * limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	var models []*Model
	for _, sm := range sms {
		models = append(models, toDomainModelFromSQL(sm))
	}
	return models, nil
}

func (r *SQLRepositoryImpl) FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error) {
	var sms []*sqlModel
	err := r.db.NewSelect().
		Model(&sms).
		Where("status_page_id = ? AND active = ?", statusPageID, true).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	var models []*Model
	for _, sm := range sms {
		models = append(models, toDomainModelFromSQL(sm))
	}
	return models, nil
}

func (r *SQLRepositoryImpl) FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error) {
	var sms []*sqlModel
	err := r.db.NewSelect().
		Model(&sms).
		Join("JOIN status_pages sp ON sp.id = i.status_page_id").
		Where("sp.slug = ? AND sp.published = ?", slug, true).
		Order("i.created_at DESC").
		Limit(limit).
		Offset(page * limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	var models []*Model
	for _, sm := range sms {
		models = append(models, toDomainModelFromSQL(sm))
	}
	return models, nil
}

func (r *SQLRepositoryImpl) Update(ctx context.Context, id string, entity *UpdateDto) (*Model, error) {
	query := r.db.NewUpdate().Model((*sqlModel)(nil)).Where("id = ?", id)

	hasUpdates := false

	if entity.Title != nil {
		query = query.Set("title = ?", *entity.Title)
		hasUpdates = true
	}
	if entity.Content != nil {
		query = query.Set("content = ?", *entity.Content)
		hasUpdates = true
	}
	if entity.Style != nil {
		query = query.Set("style = ?", *entity.Style)
		hasUpdates = true
	}
	if entity.Active != nil {
		query = query.Set("active = ?", *entity.Active)
		hasUpdates = true
		if *entity.Active {
			// reactivating: clear resolved_at
			query = query.Set("resolved_at = NULL")
		} else {
			// deactivating: set resolved_at
			now := time.Now()
			query = query.Set("resolved_at = ?", now)
		}
	}

	if !hasUpdates {
		return r.FindByID(ctx, id)
	}

	query = query.Set("updated_at = ?", time.Now())

	_, err := query.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *SQLRepositoryImpl) Resolve(ctx context.Context, id string) (*Model, error) {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*sqlModel)(nil)).
		Set("active = ?", false).
		Set("resolved_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *SQLRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*sqlModel)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *SQLRepositoryImpl) DeleteByStatusPageID(ctx context.Context, statusPageID string) error {
	_, err := r.db.NewDelete().Model((*sqlModel)(nil)).Where("status_page_id = ?", statusPageID).Exec(ctx)
	return err
}
