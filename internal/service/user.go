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

	candidate, err := u.userStore.GetByUsername(ctx, user.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username %w", err)
	}
	if candidate != nil {
		logger.Info().Interface("candidate", candidate).Msg("user already exists")
		return ErrUserAlreadyExists
	}

	err = u.userStore.Create(ctx, user)
	if err != nil {
		logger.Error().Err(err).Msg("create user")
		return fmt.Errorf("create user: %w", err)
	}

	logger.Info().Msg("user successfully created")
	return nil
}

func (u userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	logger := u.logger

	user, err := u.userStore.GetByUsername(ctx, username)
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		logger.Info().Str("username", username).Msg("user not found")
		return nil, ErrUserNotFound
	}

	logger.Info().Interface("user", user).Msg("successfully got user")
	return user, nil
}
