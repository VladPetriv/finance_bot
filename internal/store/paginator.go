package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/VladPetriv/finance_bot/internal/service"
)

func applyLimitAndOffsetForStatement(stmt sq.SelectBuilder, paginationFilter *service.Pagination) sq.SelectBuilder {
	if paginationFilter == nil {
		return stmt
	}

	var offset uint64
	if paginationFilter.Page > 1 {
		offset = uint64(paginationFilter.Page*paginationFilter.Limit) - uint64(paginationFilter.Limit)
	}

	return stmt.
		Limit(uint64(paginationFilter.Limit)).
		Offset(offset)
}
