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
	logger.Debug().Interface("category", category).Msg("got args")

	candidate, err := c.categoryStore.GetByTitle(ctx, category.Title)
	if err != nil {
		logger.Error().Err(err).Msg("get category from store")
		return fmt.Errorf("get category from store: %w", err)
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

	logger.Info().Interface("category", category).Msg("category created")
	return nil
}

func (c categoryService) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	logger := c.logger
	logger.Debug().Interface("userID", userID).Msg("got args")

	categories, err := c.categoryStore.GetAll(ctx, &GetALlCategoriesFilter{
		UserID: &userID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all categories from store")
		return nil, fmt.Errorf("get all categories from store: %w", err)
	}
	if len(categories) == 0 {
		logger.Info().Msg("categories not found")
		return nil, ErrCategoriesNotFound
	}

	logger.Info().Interface("categories", categories).Msg("got categories")
	return categories, nil
}
