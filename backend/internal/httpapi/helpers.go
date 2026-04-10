package httpapi

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/store"
)

var validStatuses = map[string]struct{}{
	"todo":        {},
	"in_progress": {},
	"done":        {},
}

var validPriorities = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
}

func parsePagination(c *gin.Context) (store.Pagination, map[string]string) {
	page := 1
	limit := 20
	fields := map[string]string{}

	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsed, err := parseInt(rawPage)
		if err != nil || parsed < 1 {
			fields["page"] = "must be a positive integer"
		} else {
			page = parsed
		}
	}
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := parseInt(rawLimit)
		if err != nil || parsed < 1 {
			fields["limit"] = "must be a positive integer"
		} else {
			if parsed > 100 {
				parsed = 100
			}
			limit = parsed
		}
	}

	if len(fields) > 0 {
		return store.Pagination{}, fields
	}

	return store.Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}, nil
}

func parseInt(value string) (int, error) {
	result := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, errors.New("not a number")
		}
		result = result*10 + int(ch-'0')
	}
	return result, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validStatus(value string) bool {
	_, ok := validStatuses[value]
	return ok
}

func validPriority(value string) bool {
	_, ok := validPriorities[value]
	return ok
}

func parseDate(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
