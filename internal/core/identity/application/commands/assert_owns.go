package commands

import (
	"context"
	"fmt"
)

func AssertOwns(_ context.Context, userID, _ string) error {
	if userID == "" {
		return fmt.Errorf("unauthenticated")
	}
	return nil
}
