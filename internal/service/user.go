package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type userService struct {
	logger    *logger.Logger
	userStore UserStore
}

var _ UserService = (*userService)(nil)

// NewUser returns new instance of user service.
func NewUser(logger *logger.Logger, userStore UserStore) *userService {
	return &userService{
		logger:    logger,
		userStore: userStore,
	}
}

func (u userService) CreateUser(ctx context.Context, user *models.User) error {
	logger := u.logger
	logger.Debug().Interface("user", user).Msg("got args")

	candidate, err := u.userStore.Get(ctx, GetUserFilter{
		Username: user.Username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}
	if candidate != nil {
		logger.Info().Interface("candidate", candidate).Msg("user already exists")
		return ErrUserAlreadyExists
	}

	err = u.userStore.Create(ctx, user)
	if err != nil {
		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	logger.Info().Msg("user created")
	return nil
}

func (u userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	logger := u.logger
	logger.Debug().Interface("username", username).Msg("got args")

	user, err := u.userStore.Get(ctx, GetUserFilter{
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return nil, fmt.Errorf("get user from store: %w", err)
	}
	if user == nil {
		logger.Info().Msg("user not found")
		return nil, ErrUserNotFound
	}

	logger.Info().Interface("user", user).Msg("got user")
	return user, nil
}
