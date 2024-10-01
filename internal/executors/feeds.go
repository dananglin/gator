package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Feeds(s *state.State, _ Executor) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get the feeds from the database: %w", err)
	}

	fmt.Printf("Feeds:\n\n")

	for _, feed := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf(
				"unable to get the creator of %s: %w",
				feed.Name,
				err,
			)
		}

		fmt.Printf(
			"- Name: %s\n  URL: %s\n  Created by: %s\n",
			feed.Name,
			feed.Url,
			user.Name,
		)
	}

	return nil
}
