package executors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
	"github.com/google/uuid"
)

func Follow(s *state.State, exe Executor, user database.User) error {
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

	timestamp := time.Now()

	args := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	followRecord, err := s.DB.CreateFeedFollow(context.Background(), args)
	if err != nil {
		if uniqueViolation(err) {
			return errors.New("you are already following this feed")
		}

		return fmt.Errorf("unable to create the feed follow record in the database: %w", err)
	}

	fmt.Printf("You are now following the feed %q.\n", followRecord.FeedName)
	fmt.Println("DEBUG:", followRecord)

	return nil
}
