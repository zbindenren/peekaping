package incident

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, entity *CreateDto) (*Model, error) {
	args := m.Called(ctx, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockRepository) FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error) {
	args := m.Called(ctx, page, limit, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Model), args.Error(1)
}

func (m *MockRepository) FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error) {
	args := m.Called(ctx, statusPageID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Model), args.Error(1)
}

func (m *MockRepository) FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error) {
	args := m.Called(ctx, statusPageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Model), args.Error(1)
}

func (m *MockRepository) FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error) {
	args := m.Called(ctx, slug, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Model), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id string, entity *UpdateDto) (*Model, error) {
	args := m.Called(ctx, id, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockRepository) Resolve(ctx context.Context, id string) (*Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteByStatusPageID(ctx context.Context, statusPageID string) error {
	args := m.Called(ctx, statusPageID)
	return args.Error(0)
}

func setupIncidentService() (*ServiceImpl, *MockRepository) {
	mockRepo := &MockRepository{}
	logger := zap.NewNop().Sugar()

	service := NewService(mockRepo, logger).(*ServiceImpl)

	return service, mockRepo
}

func TestIncidentService_Create(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		dto := &CreateDto{
			StatusPageID: "sp-123",
			Title:        "Test Incident",
			Content:      "Something is broken",
			Style:        "warning",
		}

		expected := &Model{
			ID:           "test-id",
			StatusPageID: "sp-123",
			Title:        "Test Incident",
			Content:      "Something is broken",
			Style:        "warning",
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockRepo.On("Create", ctx, dto).Return(expected, nil)

		result, err := service.Create(ctx, dto)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		dto := &CreateDto{
			StatusPageID: "sp-456",
			Title:        "Another Incident",
			Style:        "danger",
		}

		mockRepo.On("Create", ctx, dto).Return(nil, errors.New("repository error"))

		result, err := service.Create(ctx, dto)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_FindByID(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful find", func(t *testing.T) {
		expected := &Model{
			ID:           "test-id",
			StatusPageID: "sp-123",
			Title:        "Test Incident",
		}

		mockRepo.On("FindByID", ctx, "test-id").Return(expected, nil)

		result, err := service.FindByID(ctx, "test-id")

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found returns nil", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "nonexistent").Return(nil, nil)

		result, err := service.FindByID(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "error-id").Return(nil, errors.New("repository error"))

		result, err := service.FindByID(ctx, "error-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_FindAll(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful find all", func(t *testing.T) {
		expected := []*Model{
			{ID: "incident-1", Title: "Incident 1"},
			{ID: "incident-2", Title: "Incident 2"},
		}

		mockRepo.On("FindAll", ctx, 0, 10, "test").Return(expected, nil)

		result, err := service.FindAll(ctx, 0, 10, "test")

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("FindAll", ctx, 1, 10, "").Return(nil, errors.New("repository error"))

		result, err := service.FindAll(ctx, 1, 10, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_FindByStatusPageID(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful find by status page ID", func(t *testing.T) {
		expected := []*Model{
			{ID: "incident-1", StatusPageID: "sp-123"},
			{ID: "incident-2", StatusPageID: "sp-123"},
		}

		mockRepo.On("FindByStatusPageID", ctx, "sp-123", 0, 10).Return(expected, nil)

		result, err := service.FindByStatusPageID(ctx, "sp-123", 0, 10)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("FindByStatusPageID", ctx, "sp-456", 0, 10).Return(nil, errors.New("repository error"))

		result, err := service.FindByStatusPageID(ctx, "sp-456", 0, 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_FindActiveByStatusPageID(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful find active", func(t *testing.T) {
		expected := []*Model{
			{ID: "incident-1", StatusPageID: "sp-123", Active: true},
		}

		mockRepo.On("FindActiveByStatusPageID", ctx, "sp-123").Return(expected, nil)

		result, err := service.FindActiveByStatusPageID(ctx, "sp-123")

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("FindActiveByStatusPageID", ctx, "sp-456").Return(nil, errors.New("repository error"))

		result, err := service.FindActiveByStatusPageID(ctx, "sp-456")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_FindByStatusPageSlug(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful find by slug", func(t *testing.T) {
		expected := []*Model{
			{ID: "incident-1", StatusPageID: "sp-123"},
		}

		mockRepo.On("FindByStatusPageSlug", ctx, "my-page", 0, 10).Return(expected, nil)

		result, err := service.FindByStatusPageSlug(ctx, "my-page", 0, 10)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("FindByStatusPageSlug", ctx, "bad-slug", 0, 10).Return(nil, errors.New("repository error"))

		result, err := service.FindByStatusPageSlug(ctx, "bad-slug", 0, 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		service, mockRepo := setupIncidentService()
		title := "Updated Title"
		dto := &UpdateDto{Title: &title}
		expected := &Model{
			ID:    "test-id",
			Title: title,
		}

		mockRepo.On("Update", ctx, "test-id", dto).Return(expected, nil)

		result, err := service.Update(ctx, "test-id", dto)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		service, mockRepo := setupIncidentService()
		title := "Updated"
		dto := &UpdateDto{Title: &title}

		mockRepo.On("Update", ctx, "test-id", dto).Return(nil, errors.New("update failed"))

		result, err := service.Update(ctx, "test-id", dto)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "update failed")
		mockRepo.AssertExpectations(t)
	})

	t.Run("reactivation clears resolved_at", func(t *testing.T) {
		service, mockRepo := setupIncidentService()
		active := true
		dto := &UpdateDto{Active: &active}
		expected := &Model{
			ID:         "test-id",
			Active:     true,
			ResolvedAt: nil,
		}

		mockRepo.On("Update", ctx, "test-id", dto).Return(expected, nil)

		result, err := service.Update(ctx, "test-id", dto)

		assert.NoError(t, err)
		assert.True(t, result.Active)
		assert.Nil(t, result.ResolvedAt)
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_Resolve(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful resolve", func(t *testing.T) {
		now := time.Now()
		expected := &Model{
			ID:         "test-id",
			Active:     false,
			ResolvedAt: &now,
		}

		mockRepo.On("Resolve", ctx, "test-id").Return(expected, nil)

		result, err := service.Resolve(ctx, "test-id")

		assert.NoError(t, err)
		assert.False(t, result.Active)
		assert.NotNil(t, result.ResolvedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("Resolve", ctx, "error-id").Return(nil, errors.New("resolve failed"))

		result, err := service.Resolve(ctx, "error-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "resolve failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_Delete(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "test-id").Return(nil)

		err := service.Delete(ctx, "test-id")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "error-id").Return(errors.New("delete failed"))

		err := service.Delete(ctx, "error-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestIncidentService_DeleteByStatusPageID(t *testing.T) {
	service, mockRepo := setupIncidentService()
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.On("DeleteByStatusPageID", ctx, "sp-123").Return(nil)

		err := service.DeleteByStatusPageID(ctx, "sp-123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("DeleteByStatusPageID", ctx, "sp-456").Return(errors.New("delete failed"))

		err := service.DeleteByStatusPageID(ctx, "sp-456")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestNewService(t *testing.T) {
	mockRepo := &MockRepository{}
	logger := zap.NewNop().Sugar()

	service := NewService(mockRepo, logger)

	assert.NotNil(t, service)
	assert.IsType(t, &ServiceImpl{}, service)

	serviceImpl := service.(*ServiceImpl)
	assert.Equal(t, mockRepo, serviceImpl.repository)
	assert.NotNil(t, serviceImpl.logger)
}
