package validate

import (
	"strings"

	"github.com/target/goalert/validation"
)

// SID will validate an SID, returning a FieldError if invalid.
func SID(fname, value string) error {

	if !strings.HasPrefix(value, "MG") {
		return validation.NewFieldError(fname, "must begin with MG")
	}

	return nil
}