package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/VladPetriv/finance_bot/pkg/database"
)

type categoryStore struct {
	*database.PostgreSQL
}

// NewCategory returns a new instance of category store.
func NewCategory(db *database.PostgreSQL) *categoryStore {
	return &categoryStore{
		db,
	}
}
func (c *categoryStore) Create(ctx context.Context, category *models.Category) error {
	_, err := c.DB.ExecContext(ctx,
		"INSERT INTO categories (id, user_id, title) VALUES ($1, $2, $3);",
		category.ID, category.UserID, category.Title,
	)

	return err
}

func (c *categoryStore) Get(ctx context.Context, filter service.GetCategoryFilter) (*models.Category, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "user_id", "title").
		From("categories")

	if filter.ID != "" {
		stmt = stmt.Where(sq.Eq{"id": filter.ID})
	}
	if filter.Title != "" {
		stmt = stmt.Where(sq.Eq{"title": filter.Title})
	}
	if filter.UserID != "" {
		stmt = stmt.Where(sq.Eq{"user_id": filter.UserID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get category query: %w", err)
	}

	var category models.Category
	err = c.DB.GetContext(ctx, &category, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (c *categoryStore) List(ctx context.Context, filter *service.ListCategoriesFilter) ([]models.Category, error) {
	stmt := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "user_id", "title").
		From("categories")

	if filter.UserID != "" {
		stmt = stmt.Where(sq.Eq{"user_id": filter.UserID})
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list categories query: %w", err)
	}

	var categories []models.Category
	err = c.DB.SelectContext(ctx, &categories, query, args...)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (c *categoryStore) Update(ctx context.Context, category *models.Category) error {
	_, err := c.DB.ExecContext(
		ctx,
		"UPDATE categories SET user_id = $2, title = $3 WHERE id = $1;",
		category.ID, category.UserID, category.Title,
	)
	return err
}

func (c *categoryStore) Delete(ctx context.Context, categoryID string) error {
	_, err := c.DB.ExecContext(ctx, "DELETE FROM categories WHERE id = $1;", categoryID)
	return err
}
