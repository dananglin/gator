package executors

import (
	"context"
	"fmt"
	"strconv"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func Browse(s *state.State, exe Executor, user database.User) error {
	if len(exe.Args) > 1 {
		return fmt.Errorf("unexpected number of arguments: want 0 or 1, got %d", len(exe.Args))
	}

	var err error

	limit := 2

	if len(exe.Args) == 1 {
		limit, err = strconv.Atoi(exe.Args[0])
		if err != nil {
			return fmt.Errorf("unable to convert %s to a number: %w", exe.Args[0], err)
		}
	}

	args := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.DB.GetPostsForUser(context.Background(), args)
	if err != nil {
		return fmt.Errorf("unable to get the posts: %w", err)
	}

	fmt.Printf("\nPosts:\n\n")

	for _, post := range posts {
		fmt.Printf(
			"- Title: %s\n  URL: %s\n  Published at: %s\n",
			post.Title,
			post.Url,
			post.PublishedAt,
		)
	}

	return nil
}
