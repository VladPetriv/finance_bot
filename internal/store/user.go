package store

import (
	"context"
	"errors"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userStore struct {
	*database.MongoDB
}

var _ service.UserStore = (*userStore)(nil)

var collectionUser = "User"

// NewUserStore returns new instance of user store.
func NewUserStore(db *database.MongoDB) *userStore {
	return &userStore{
		db,
	}
}

func (u userStore) Create(ctx context.Context, user *models.User) error {
	_, err := u.DB.Collection(collectionUser).InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (u userStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := u.DB.Collection(collectionUser).FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}
