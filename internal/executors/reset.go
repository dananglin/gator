package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Reset(s *state.State, _ Executor) error {
	if err := s.DB.DeleteAllUsers(context.Background()); err != nil {
		fmt.Errorf("unable to delete the users from the database: %w", err)
	}

	fmt.Println("Successfully removed all users from the database.")

	return nil
}
