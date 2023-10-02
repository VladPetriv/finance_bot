// Code generated by mockery v2.34.2. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/VladPetriv/finance_bot/internal/models"
	mock "github.com/stretchr/testify/mock"

	service "github.com/VladPetriv/finance_bot/internal/service"
)

// CategoryStore is an autogenerated mock type for the CategoryStore type
type CategoryStore struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, category
func (_m *CategoryStore) Create(ctx context.Context, category *models.Category) error {
	ret := _m.Called(ctx, category)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Category) error); ok {
		r0 = rf(ctx, category)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, categoryID
func (_m *CategoryStore) Delete(ctx context.Context, categoryID string) error {
	ret := _m.Called(ctx, categoryID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, categoryID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, filter
func (_m *CategoryStore) Get(ctx context.Context, filter service.GetCategoryFilter) (*models.Category, error) {
	ret := _m.Called(ctx, filter)

	var r0 *models.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, service.GetCategoryFilter) (*models.Category, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, service.GetCategoryFilter) *models.Category); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, service.GetCategoryFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: ctx, filters
func (_m *CategoryStore) GetAll(ctx context.Context, filters *service.GetALlCategoriesFilter) ([]models.Category, error) {
	ret := _m.Called(ctx, filters)

	var r0 []models.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *service.GetALlCategoriesFilter) ([]models.Category, error)); ok {
		return rf(ctx, filters)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *service.GetALlCategoriesFilter) []models.Category); ok {
		r0 = rf(ctx, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *service.GetALlCategoriesFilter) error); ok {
		r1 = rf(ctx, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewCategoryStore creates a new instance of CategoryStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCategoryStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *CategoryStore {
	mock := &CategoryStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
