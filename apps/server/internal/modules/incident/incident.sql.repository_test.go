package incident

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func setupTestDB(t *testing.T) *bun.DB {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	_, err = db.Exec(`
		CREATE TABLE incidents (
			id TEXT PRIMARY KEY,
			status_page_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			style TEXT NOT NULL DEFAULT 'warning',
			active BOOLEAN NOT NULL DEFAULT TRUE,
			resolved_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestSQLRepositoryImpl_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	t.Run("successful creation", func(t *testing.T) {
		dto := &CreateDto{
			StatusPageID: "sp-123",
			Title:        "Test Incident",
			Content:      "Something is broken",
			Style:        "warning",
		}

		result, err := repo.Create(ctx, dto)

		require.NoError(t, err)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, dto.StatusPageID, result.StatusPageID)
		assert.Equal(t, dto.Title, result.Title)
		assert.Equal(t, dto.Content, result.Content)
		assert.Equal(t, dto.Style, result.Style)
		assert.True(t, result.Active)
		assert.Nil(t, result.ResolvedAt)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})
}

func TestSQLRepositoryImpl_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	t.Run("successful find", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-123",
			Title:        "Find Me",
			Style:        "info",
		})
		require.NoError(t, err)

		result, err := repo.FindByID(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "Find Me", result.Title)
	})

	t.Run("not found returns nil", func(t *testing.T) {
		result, err := repo.FindByID(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestSQLRepositoryImpl_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	// seed data
	_, err := repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Alpha Incident", Style: "info"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Beta Incident", Style: "warning"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-2", Title: "Gamma Incident", Style: "danger"})
	require.NoError(t, err)

	t.Run("find all without filter", func(t *testing.T) {
		results, err := repo.FindAll(ctx, 0, 10, "")

		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("find all with search query", func(t *testing.T) {
		results, err := repo.FindAll(ctx, 0, 10, "alpha")

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Alpha Incident", results[0].Title)
	})

	t.Run("pagination", func(t *testing.T) {
		results, err := repo.FindAll(ctx, 0, 2, "")

		require.NoError(t, err)
		assert.Len(t, results, 2)

		results2, err := repo.FindAll(ctx, 1, 2, "")

		require.NoError(t, err)
		assert.Len(t, results2, 1)
	})

	t.Run("ordered by created_at DESC", func(t *testing.T) {
		results, err := repo.FindAll(ctx, 0, 10, "")

		require.NoError(t, err)
		require.Len(t, results, 3)
		// newest first
		assert.Equal(t, "Gamma Incident", results[0].Title)
		assert.Equal(t, "Beta Incident", results[1].Title)
		assert.Equal(t, "Alpha Incident", results[2].Title)
	})
}

func TestSQLRepositoryImpl_FindByStatusPageID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	_, err := repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Incident A", Style: "info"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Incident B", Style: "warning"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-2", Title: "Incident C", Style: "danger"})
	require.NoError(t, err)

	t.Run("filters by status page ID", func(t *testing.T) {
		results, err := repo.FindByStatusPageID(ctx, "sp-1", 0, 10)

		require.NoError(t, err)
		assert.Len(t, results, 2)
		for _, r := range results {
			assert.Equal(t, "sp-1", r.StatusPageID)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		results, err := repo.FindByStatusPageID(ctx, "sp-1", 0, 1)

		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("no results for unknown status page", func(t *testing.T) {
		results, err := repo.FindByStatusPageID(ctx, "sp-unknown", 0, 10)

		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestSQLRepositoryImpl_FindActiveByStatusPageID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	_, err := repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Active Incident", Style: "warning"})
	require.NoError(t, err)
	created, err := repo.Create(ctx, &CreateDto{StatusPageID: "sp-1", Title: "Resolved Incident", Style: "info"})
	require.NoError(t, err)

	// resolve the second incident
	_, err = repo.Resolve(ctx, created.ID)
	require.NoError(t, err)

	t.Run("returns only active incidents", func(t *testing.T) {
		results, err := repo.FindActiveByStatusPageID(ctx, "sp-1")

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Active Incident", results[0].Title)
		assert.True(t, results[0].Active)
	})

	t.Run("no active incidents", func(t *testing.T) {
		results, err := repo.FindActiveByStatusPageID(ctx, "sp-empty")

		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestSQLRepositoryImpl_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	t.Run("update title", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "Original Title",
			Style:        "info",
		})
		require.NoError(t, err)

		newTitle := "Updated Title"
		result, err := repo.Update(ctx, created.ID, &UpdateDto{Title: &newTitle})

		require.NoError(t, err)
		assert.Equal(t, "Updated Title", result.Title)
	})

	t.Run("deactivate sets resolved_at", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "To Deactivate",
			Style:        "warning",
		})
		require.NoError(t, err)

		active := false
		result, err := repo.Update(ctx, created.ID, &UpdateDto{Active: &active})

		require.NoError(t, err)
		assert.False(t, result.Active)
		assert.NotNil(t, result.ResolvedAt)
	})

	t.Run("reactivate clears resolved_at", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "To Reactivate",
			Style:        "danger",
		})
		require.NoError(t, err)

		// first deactivate
		active := false
		_, err = repo.Update(ctx, created.ID, &UpdateDto{Active: &active})
		require.NoError(t, err)

		// then reactivate
		active = true
		result, err := repo.Update(ctx, created.ID, &UpdateDto{Active: &active})

		require.NoError(t, err)
		assert.True(t, result.Active)
		assert.Nil(t, result.ResolvedAt)
	})

	t.Run("no updates returns current model", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "No Changes",
			Style:        "info",
		})
		require.NoError(t, err)

		result, err := repo.Update(ctx, created.ID, &UpdateDto{})

		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "No Changes", result.Title)
	})
}

func TestSQLRepositoryImpl_Resolve(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	t.Run("successful resolve", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "To Resolve",
			Style:        "warning",
		})
		require.NoError(t, err)
		assert.True(t, created.Active)

		result, err := repo.Resolve(ctx, created.ID)

		require.NoError(t, err)
		assert.False(t, result.Active)
		assert.NotNil(t, result.ResolvedAt)
	})
}

func TestSQLRepositoryImpl_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	t.Run("successful deletion", func(t *testing.T) {
		created, err := repo.Create(ctx, &CreateDto{
			StatusPageID: "sp-1",
			Title:        "To Delete",
			Style:        "info",
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, created.ID)

		assert.NoError(t, err)

		result, err := repo.FindByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete nonexistent is not an error", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent")
		assert.NoError(t, err)
	})
}

func TestSQLRepositoryImpl_DeleteByStatusPageID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLRepository(db)
	ctx := t.Context()

	_, err := repo.Create(ctx, &CreateDto{StatusPageID: "sp-del", Title: "Inc 1", Style: "info"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-del", Title: "Inc 2", Style: "warning"})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &CreateDto{StatusPageID: "sp-keep", Title: "Inc 3", Style: "danger"})
	require.NoError(t, err)

	t.Run("deletes all incidents for status page", func(t *testing.T) {
		err := repo.DeleteByStatusPageID(ctx, "sp-del")

		assert.NoError(t, err)

		remaining, err := repo.FindAll(ctx, 0, 10, "")
		require.NoError(t, err)
		assert.Len(t, remaining, 1)
		assert.Equal(t, "sp-keep", remaining[0].StatusPageID)
	})
}
