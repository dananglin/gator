package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Following(s *state.State, _ Executor, user database.User) error {
	following, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("unable to get the list of feeds from the database: %w", err)
	}

	if len(following) == 0 {
		fmt.Println("You are not following any feeds.")

		return nil
	}

	fmt.Printf("\nYou are following:\n\n")

	for _, feed := range following {
		fmt.Printf("- %s\n", feed)
	}

	return nil
}
