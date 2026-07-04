package httptransport

import (
	"strconv"

	"sakeofher/internal/domain"
)

func pathInt64(r interface{ PathValue(string) string }, name string) (int64, error) {
	value := r.PathValue(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, domain.ErrInvalidInput
	}

	return id, nil
}
