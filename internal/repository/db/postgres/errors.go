package pg

import (
	"errors"

	"github.com/lib/pq"
)

func IsDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// PostgreSQL error code 23505 = unique_violation
		return pqErr.Code == "23505"
	}
	return false
}
