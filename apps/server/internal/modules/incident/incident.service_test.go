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

func createTestService() (*ServiceImpl, *MockRepository) {
	mockRepo := &MockRepository{}
	logger := zap.NewNop().Sugar()

	service := &ServiceImpl{
		repository: mockRepo,
		logger:     logger,
	}

	return service, mockRepo
}

func createTestCreateDto() *CreateDto {
	return &CreateDto{
		StatusPageID: "sp-123",
		Title:        "Test Incident",
		Content:      "Something is broken",
		Style:        "warning",
	}
}

func createTestModel() *Model {
	return &Model{
		ID:           "test-id",
		StatusPageID: "sp-123",
		Title:        "Test Incident",
		Content:      "Something is broken",
		Style:        "warning",
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestServiceImpl_Create_Success(t *testing.T) {
	service, mockRepo := createTestService()
	dto := createTestCreateDto()
	expected := createTestModel()

	mockRepo.On("Create", mock.Anything, dto).Return(expected, nil)

	result, err := service.Create(t.Context(), dto)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Create_RepositoryError(t *testing.T) {
	service, mockRepo := createTestService()
	dto := createTestCreateDto()
	expectedErr := errors.New("repository error")

	mockRepo.On("Create", mock.Anything, dto).Return(nil, expectedErr)

	result, err := service.Create(t.Context(), dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_FindByID_Success(t *testing.T) {
	service, mockRepo := createTestService()
	expected := createTestModel()

	mockRepo.On("FindByID", mock.Anything, "test-id").Return(expected, nil)

	result, err := service.FindByID(t.Context(), "test-id")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_FindByID_NotFound(t *testing.T) {
	service, mockRepo := createTestService()

	mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	result, err := service.FindByID(t.Context(), "nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_FindActiveByStatusPageID_Success(t *testing.T) {
	service, mockRepo := createTestService()
	expected := []*Model{createTestModel()}

	mockRepo.On("FindActiveByStatusPageID", mock.Anything, "sp-123").Return(expected, nil)

	result, err := service.FindActiveByStatusPageID(t.Context(), "sp-123")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Update_Success(t *testing.T) {
	service, mockRepo := createTestService()
	title := "Updated Title"
	dto := &UpdateDto{Title: &title}
	expected := createTestModel()
	expected.Title = title

	mockRepo.On("Update", mock.Anything, "test-id", dto).Return(expected, nil)

	result, err := service.Update(t.Context(), "test-id", dto)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Update_RepositoryError(t *testing.T) {
	service, mockRepo := createTestService()
	title := "Updated"
	dto := &UpdateDto{Title: &title}
	expectedErr := errors.New("update failed")

	mockRepo.On("Update", mock.Anything, "test-id", dto).Return(nil, expectedErr)

	result, err := service.Update(t.Context(), "test-id", dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Update_Reactivate(t *testing.T) {
	service, mockRepo := createTestService()
	active := true
	dto := &UpdateDto{Active: &active}
	expected := createTestModel()
	expected.Active = true
	expected.ResolvedAt = nil

	mockRepo.On("Update", mock.Anything, "test-id", dto).Return(expected, nil)

	result, err := service.Update(t.Context(), "test-id", dto)

	assert.NoError(t, err)
	assert.True(t, result.Active)
	assert.Nil(t, result.ResolvedAt)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Resolve_Success(t *testing.T) {
	service, mockRepo := createTestService()
	now := time.Now()
	expected := createTestModel()
	expected.Active = false
	expected.ResolvedAt = &now

	mockRepo.On("Resolve", mock.Anything, "test-id").Return(expected, nil)

	result, err := service.Resolve(t.Context(), "test-id")

	assert.NoError(t, err)
	assert.False(t, result.Active)
	assert.NotNil(t, result.ResolvedAt)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Delete_Success(t *testing.T) {
	service, mockRepo := createTestService()

	mockRepo.On("Delete", mock.Anything, "test-id").Return(nil)

	err := service.Delete(t.Context(), "test-id")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_Delete_RepositoryError(t *testing.T) {
	service, mockRepo := createTestService()
	expectedErr := errors.New("delete failed")

	mockRepo.On("Delete", mock.Anything, "test-id").Return(expectedErr)

	err := service.Delete(t.Context(), "test-id")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceImpl_DeleteByStatusPageID_Success(t *testing.T) {
	service, mockRepo := createTestService()

	mockRepo.On("DeleteByStatusPageID", mock.Anything, "sp-123").Return(nil)

	err := service.DeleteByStatusPageID(t.Context(), "sp-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
