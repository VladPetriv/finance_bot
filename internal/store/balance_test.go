package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/VladPetriv/finance_bot/internal/store"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestBalance_Get(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	db, err := database.NewMongoDB("mongodb: //localhost:27017", "api")
	require.NoError(t, err)

	balanceStore := store.NewBalance(db)

	tests := []struct {
		name  string
		input string
	}{}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := balanceStore.Get(ctx, tt.input)
			fmt.Println(got, err)
		})
	}
}
