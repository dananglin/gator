package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Unfollow(s *state.State, exe Executor, user database.User) error {
	wantNumArgs := 1

	if len(exe.Args) != wantNumArgs {
		return fmt.Errorf(
			"unexpected number of arguments: want %d, got %d",
			wantNumArgs,
			len(exe.Args),
		)
	}

	url := exe.Args[0]

	feed, err := s.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("unable to get the feed data from the database: %w", err)
	}

	args := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	if err := s.DB.DeleteFeedFollow(context.Background(), args); err != nil {
		return fmt.Errorf("unable to delete the feed follow record from the database: %w", err)
	}

	fmt.Printf("You have successfully unfollowed %q.\n", feed.Name)

	return nil
}
