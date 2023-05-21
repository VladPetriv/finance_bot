package service

import (
	"context"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type categoryService struct {
	logger        *logger.Logger
	categoryStore CategoryStore
}

var _ CategoryService = (*categoryService)(nil)

// NewCategory returns new instance of category service.
func NewCategory(logger *logger.Logger, categoryStore CategoryStore) *categoryService {
	return &categoryService{
		logger:        logger,
		categoryStore: categoryStore,
	}
}

func (c categoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	logger := c.logger

	candidate, err := c.categoryStore.GetByTitle(ctx, category.Title)
	if err != nil {
		logger.Error().Err(err).Msg("get category by title")
		return fmt.Errorf("get category by title: %w", err)
	}
	if candidate != nil {
		logger.Info().Interface("candidate", candidate).Msgf("category with %s title already exist", category.Title)
		return ErrCategoryAlreadyExists
	}

	err = c.categoryStore.Create(ctx, category)
	if err != nil {
		logger.Error().Err(err).Msg("create category")
		return fmt.Errorf("create category: %w", err)
	}

	logger.Info().Interface("category", category).Msg("category successfully created")
	return nil
}
