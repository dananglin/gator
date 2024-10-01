package executors

import (
	"errors"

	"github.com/lib/pq"
)

func uniqueViolation(err error) bool {
	var pqError *pq.Error

	if errors.As(err, &pqError) {
		if pqError.Code.Name() == "unique_violation" {
			return true
		}
	}

	return false
}
