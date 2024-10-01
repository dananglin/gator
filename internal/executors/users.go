package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Users(s *state.State, _ Executor) error {
	users, err := s.DB.GetAllUsers(context.Background())
	if err != nil {
		fmt.Errorf("unable to get the users from the database: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("There are no registered users.")

		return nil
	}

	fmt.Printf("Registered users:\n\n")

	for _, user := range users {
		if user.Name == s.Config.CurrentUsername {
			fmt.Printf("- %s (current)\n", user.Name)
		} else {
			fmt.Printf("- %s\n", user.Name)
		}
	}

	return nil
}
