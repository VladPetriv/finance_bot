package store_test

import (
	"context"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	amount3000 = float32(3000)
	amount4000 = float32(4000)
)

func TestBalance_Get(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	balanceStore := store.NewBalance(db)

	balanceID := uuid.NewString()

	err = balanceStore.Create(ctx, &models.Balance{
		ID:     balanceID,
		Amount: &amount3000,
	}) //
	require.NoError(t, err)

	tests := []struct {
		desc  string
		input string
		want  *models.Balance
	}{
		{
			desc:  "should return balance by id",
			input: balanceID,
			want: &models.Balance{
				ID:     balanceID,
				Amount: &amount3000,
			},
		},
		{
			desc:  "should not return balance by id because it's not exist",
			input: uuid.NewString(),
			want:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			got, err := balanceStore.Get(ctx, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			t.Cleanup(func() {
				err := balanceStore.Delete(ctx, balanceID)
				require.NoError(t, err)
			})
		})
	}
}

func TestBalance_Create(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	balanceStore := store.NewBalance(db)
	balanceID := uuid.NewString()

	tests := []struct {
		desc          string
		input         *models.Balance
		expectedError error
	}{
		{
			desc: "should create new balance",
			input: &models.Balance{
				ID:     balanceID,
				Amount: &amount3000,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			err := balanceStore.Create(ctx, tt.input)
			assert.NoError(t, err)

			t.Cleanup(func() {
				err := balanceStore.Delete(ctx, balanceID)
				require.NoError(t, err)
			})
		})
	}
}

func TestBalance_Update(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	balanceStore := store.NewBalance(db)

	balanceID1 := uuid.NewString()
	balanceID2 := uuid.NewString()

	tests := []struct {
		desc          string
		input         *models.Balance
		want          *models.Balance
		expectedError error
	}{
		{
			desc: "should update existed balance",
			input: &models.Balance{
				ID:     balanceID1,
				Amount: &amount3000,
			},
			want: &models.Balance{
				ID:     balanceID1,
				Amount: &amount4000,
			},
		},
		{
			desc: "should not update balance because it's not exist",
			input: &models.Balance{
				ID:     balanceID2,
				Amount: &amount3000,
			},
			want: &models.Balance{
				ID:     uuid.NewString(),
				Amount: &amount3000,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			err = balanceStore.Create(ctx, tt.input)
			require.NoError(t, err)

			err := balanceStore.Update(ctx, tt.want)
			assert.NoError(t, err)

			if tt.input != nil {
				got, err := balanceStore.Get(ctx, tt.input.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Amount, got.Amount)
			}

			t.Cleanup(func() {
				err := balanceStore.Delete(ctx, tt.input.ID)
				require.NoError(t, err)
			})
		})
	}
}

func TestBalance_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.TODO() //nolint: forbidigo

	db, err := database.NewMongoDB(ctx, "mongodb://localhost:27017", "api")
	require.NoError(t, err)

	balanceStore := store.NewBalance(db)

	balanceID := uuid.NewString()

	tests := []struct {
		desc                string
		input               *models.Balance
		shouldCreateBalance bool
	}{
		{
			desc: "should delete existed balance",
			input: &models.Balance{
				ID:     balanceID,
				Amount: &amount3000,
			},
			shouldCreateBalance: true,
		},
		{
			desc: "should not delete balance because it's not exist",
			input: &models.Balance{
				ID: uuid.NewString(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			if tt.shouldCreateBalance {
				err = balanceStore.Create(ctx, tt.input)
				require.NoError(t, err)
			}

			err := balanceStore.Delete(ctx, tt.input.ID)
			assert.NoError(t, err)
		})
	}
}
