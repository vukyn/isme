package query

import (
	"strings"

	"github.com/uptrace/bun"
)

type Pagination struct {
	Page      int
	Size      int
	SortBy    string
	SortOrder string
	CountOnly bool
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Size
}

func (p *Pagination) GetLimit() int {
	return p.Size
}

func SelectWithPagination(query *bun.SelectQuery, paging Pagination, defaultSort string) *bun.SelectQuery {
	if paging.SortBy != "" {
		if strings.ToLower(paging.SortOrder) == "asc" {
			query = query.Order(paging.SortBy + " ASC")
		} else {
			query = query.Order(paging.SortBy + " DESC")
		}
	} else {
		query = query.Order(defaultSort)
	}

	if paging.GetLimit() > 0 {
		query = query.Limit(paging.GetLimit())
	}

	if paging.GetOffset() > 0 {
		query = query.Offset(paging.GetOffset())
	}
	return query
}

// BoolToInt converts a boolean value to integer for SQLite compatibility
// SQLite stores booleans as integers (0 for false, 1 for true)
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// BoolPtrToInt converts a boolean pointer to integer for SQLite compatibility
// Returns 0 if the pointer is nil
func BoolPtrToInt(b *bool) int {
	if b == nil {
		return 0
	}
	return BoolToInt(*b)
}
