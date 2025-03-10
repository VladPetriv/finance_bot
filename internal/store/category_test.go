package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory_Create(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	testCaseDB := createTestDB(t, "category_create")
	userStore := store.NewUser(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)

	userID := uuid.NewString()
	categoryID := uuid.NewString()
	err := userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + uuid.NewString(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc                 string
		preconditions        *models.Category
		args                 *models.Category
		expectDuplicateError bool
	}{
		{
			desc: "created category",
			args: &models.Category{
				ID:     uuid.NewString(),
				UserID: userID,
				Title:  "test_create_1",
			},
		},
		{
			desc: "category not created because already exists",
			preconditions: &models.Category{
				ID:     categoryID,
				UserID: userID,
				Title:  "test_create_2",
			},
			args: &models.Category{
				ID:     categoryID,
				UserID: userID,
				Title:  "test_create_2",
			},
			expectDuplicateError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err = categoryStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
				if tc.args != nil {
					err = categoryStore.Delete(ctx, tc.args.ID)
					assert.NoError(t, err)
				}
			})

			err := categoryStore.Create(ctx, tc.args)
			if tc.expectDuplicateError {
				assert.True(t, isDuplicateKeyError(err))
				return
			}

			assert.NoError(t, err)

			actual, err := categoryStore.Get(ctx, service.GetCategoryFilter{ID: tc.args.ID})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
			assert.Equal(t, tc.args.ID, actual.ID)
			assert.Equal(t, tc.args.UserID, actual.UserID)
			assert.Equal(t, tc.args.Title, actual.Title)
		})
	}
}

func TestCategory_Get(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	testCaseDB := createTestDB(t, "category_get")
	userStore := store.NewUser(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)

	categoryID1, categoryID2, categoryID3 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	userID1, userID2 := uuid.NewString(), uuid.NewString()
	for _, userID := range [...]string{userID1, userID2} {
		err := userStore.Create(ctx, &models.User{
			ID:       userID,
			Username: "test" + uuid.NewString(),
		})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		for _, userID := range [...]string{userID1, userID2} {
			err := deleteUserByID(testCaseDB.DB, userID)
			require.NoError(t, err)
		}
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.Category
		args          service.GetCategoryFilter
		expected      *models.Category
	}{
		{
			desc: "found category by title",
			preconditions: &models.Category{
				ID:     categoryID1,
				UserID: userID1,
				Title:  "title_get_1",
			},
			args: service.GetCategoryFilter{
				Title: "title_get_1",
			},
			expected: &models.Category{
				ID:     categoryID1,
				UserID: userID1,
				Title:  "title_get_1",
			},
		},
		{
			desc: "found category by id",
			preconditions: &models.Category{
				ID:     categoryID2,
				UserID: userID1,
				Title:  "test_get_2",
			},
			args: service.GetCategoryFilter{
				ID: categoryID2,
			},
			expected: &models.Category{
				ID:     categoryID2,
				UserID: userID1,
				Title:  "test_get_2",
			},
		},
		{
			desc: "found category by user id",
			preconditions: &models.Category{
				ID:     categoryID3,
				UserID: userID2,
				Title:  "test_get_user_id",
			},
			args: service.GetCategoryFilter{
				UserID: userID2,
			},
			expected: &models.Category{
				ID:     categoryID3,
				UserID: userID2,
				Title:  "test_get_user_id",
			},
		},
		{
			desc: "negative: category not found",
			args: service.GetCategoryFilter{
				Title: uuid.NewString(),
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := categoryStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			got, err := categoryStore.Get(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCategory_List(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	testCaseDB := createTestDB(t, "category_list")
	userStore := store.NewUser(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)

	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	userID1, userID2 := uuid.NewString(), uuid.NewString()
	err := userStore.Create(ctx, &models.User{
		ID:       userID1,
		Username: "test" + uuid.NewString(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, userID1)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions []models.Category
		args          *service.ListCategoriesFilter
		expected      []models.Category
	}{
		{
			desc: "found categories by user id",
			preconditions: []models.Category{
				{
					ID:     categoryID1,
					UserID: userID1,
					Title:  "title_get_1",
				},
				{
					ID:     categoryID2,
					UserID: userID1,
					Title:  "title_get_2",
				},
			},
			args: &service.ListCategoriesFilter{
				UserID: userID1,
			},
			expected: []models.Category{
				{
					ID:     categoryID1,
					UserID: userID1,
					Title:  "title_get_1",
				},
				{
					ID:     categoryID2,
					UserID: userID1,
					Title:  "title_get_2",
				},
			},
		},
		{
			desc: "categories not found",
			args: &service.ListCategoriesFilter{
				UserID: userID2,
			},
			expected: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			for _, category := range tc.preconditions {
				err := categoryStore.Create(ctx, &category)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, category := range tc.preconditions {
					err := categoryStore.Delete(ctx, category.ID)
					assert.NoError(t, err)
				}
			})

			got, err := categoryStore.List(ctx, tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestCategory_Update(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	testCaseDB := createTestDB(t, "category_update")
	userStore := store.NewUser(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)

	userID := uuid.NewString()
	categoryID1, categoryID2 := uuid.NewString(), uuid.NewString()
	err := userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + uuid.NewString(),
	})
	require.NoError(t, err)

	testCases := [...]struct {
		desc          string
		preconditions *models.Category
		args          *models.Category
		expected      *models.Category
	}{
		{
			desc: "category updated",
			preconditions: &models.Category{
				ID:     categoryID1,
				UserID: userID,
				Title:  "old_title",
			},
			args: &models.Category{
				ID:     categoryID1,
				UserID: userID,
				Title:  "new_title",
			},
			expected: &models.Category{
				ID:     categoryID1,
				UserID: userID,
				Title:  "new_title",
			},
		},
		{
			desc: "category not updated because of not existed id",
			preconditions: &models.Category{
				ID:     categoryID2,
				UserID: userID,
				Title:  "test_title",
			},
			args: &models.Category{
				ID:     uuid.NewString(),
				UserID: userID,
				Title:  "updated_title",
			},
			expected: &models.Category{
				ID:     categoryID2,
				UserID: userID,
				Title:  "test_title",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err = categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err = categoryStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err = categoryStore.Update(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := categoryStore.Get(ctx, service.GetCategoryFilter{ID: tc.preconditions.ID})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCategory_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo
	testCaseDB := createTestDB(t, "category_delete")
	userStore := store.NewUser(testCaseDB)
	categoryStore := store.NewCategory(testCaseDB)

	userID := uuid.NewString()
	categoryID := uuid.NewString()
	err := userStore.Create(ctx, &models.User{
		ID:       userID,
		Username: "test" + uuid.NewString(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := deleteUserByID(testCaseDB.DB, userID)
		require.NoError(t, err)
	})

	testCases := [...]struct {
		desc          string
		preconditions *models.Category
		args          string
	}{
		{
			desc: "category deleted",
			preconditions: &models.Category{
				ID:     categoryID,
				UserID: userID,
			},
			args: categoryID,
		},
		{
			desc: "category not deleted because of not existed id",
			preconditions: &models.Category{
				ID:     uuid.NewString(),
				UserID: userID,
			},
			args: uuid.NewString(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if tc.preconditions != nil {
				err := categoryStore.Create(ctx, tc.preconditions)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				if tc.preconditions != nil {
					err := categoryStore.Delete(ctx, tc.preconditions.ID)
					assert.NoError(t, err)
				}
			})

			err := categoryStore.Delete(ctx, tc.args)
			assert.NoError(t, err)

			actual, err := categoryStore.Get(ctx, service.GetCategoryFilter{ID: tc.preconditions.ID})
			assert.NoError(t, err)

			// operation should not be deleted
			if tc.preconditions.ID != tc.args {
				assert.NotNil(t, actual)
				return
			}

			assert.Nil(t, actual)
		})
	}
}
