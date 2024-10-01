package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Login(s *state.State, exe Executor) error {
	if len(exe.Args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(exe.Args))
	}

	username := exe.Args[0]

	user, err := s.DB.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("unable to get the user from the database: %w", err)
	}

	if err := s.Config.SetUser(user.Name); err != nil {
		return fmt.Errorf("login error: %w", err)
	}

	fmt.Printf("The current user is set to %q.\n", username)

	return nil
}
