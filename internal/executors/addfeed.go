package executors

import (
	"context"
	"fmt"
	"time"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
	"github.com/google/uuid"
)

func AddFeed(s *state.State, exe Executor, user database.User) error {
	wantArgs := 2

	if len(exe.Args) != wantArgs {
		return fmt.Errorf(
			"unexpected number of arguments: want %d, got %d",
			wantArgs,
			len(exe.Args),
		)
	}

	name, url := exe.Args[0], exe.Args[1]

	timestamp := time.Now()

	createdFeedArgs := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.DB.CreateFeed(context.Background(), createdFeedArgs)
	if err != nil {
		return fmt.Errorf("unable to add the feed: %w", err)
	}

	fmt.Println("Successfully added the feed.")

	fmt.Println("DEBUG:", feed)

	createFeedFollowArgs := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	followRecord, err := s.DB.CreateFeedFollow(context.Background(), createFeedFollowArgs)
	if err != nil {
		return fmt.Errorf("unable to create the feed follow record in the database: %w", err)
	}

	fmt.Printf("You are now following the feed %q.\n", followRecord.FeedName)
	fmt.Println("DEBUG:", followRecord)

	return nil
}
