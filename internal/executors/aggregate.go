package executors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/rss"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
	"github.com/google/uuid"
)

func Aggregate(s *state.State, exe Executor) error {
	if len(exe.Args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(exe.Args))
	}

	intervalArg := exe.Args[0]

	interval, err := time.ParseDuration(intervalArg)
	if err != nil {
		return fmt.Errorf("unable to parse the interval: %w", err)
	}

	fmt.Printf("Fetching feeds every %s\n", interval.String())

	tick := time.Tick(interval)

	for range tick {
		if err := scrapeFeeds(s); err != nil {
			fmt.Println("ERROR: %v", err)
		}
	}

	return nil
}

func scrapeFeeds(s *state.State) error {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get the next feed from the database: %w", err)
	}

	fmt.Printf("\nFetching feed from %s\n", feed.Url)

	feedDetails, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("unable to fetch the feed: %w", err)
	}

	timestamp := time.Now()

	lastFetched := sql.NullTime{
		Time:  timestamp,
		Valid: true,
	}

	markFeedFetchedArgs := database.MarkFeedFetchedParams{
		ID:            feed.ID,
		LastFetchedAt: lastFetched,
		UpdatedAt:     timestamp,
	}

	if err := s.DB.MarkFeedFetched(context.Background(), markFeedFetchedArgs); err != nil {
		return fmt.Errorf("unable to mark the feed as fetched in the database: %w", err)
	}

	timeParsingFormats := []string{
		time.RFC1123Z,
		time.RFC1123,
	}

	for _, item := range feedDetails.Channel.Items {
		var (
			pubDate time.Time
			err     error
		)

		pubDateFormatted := false

		for _, format := range timeParsingFormats {
			pubDate, err = time.Parse(format, item.PubDate)
			if err == nil {
				pubDateFormatted = true

				break
			}
		}

		if !pubDateFormatted {
			fmt.Printf(
				"Error: unable to format the publication date (%s) of %q.\n",
				item.PubDate,
				item.Title,
			)

			continue
		}

		timestamp := time.Now()

		args := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			FeedID:      feed.ID,
			PublishedAt: pubDate,
		}

		_, err = s.DB.CreatePost(context.Background(), args)
		if err != nil && !uniqueViolation(err) {
			fmt.Printf(
				"Error: unable to add the post %q to the database: %v.\n",
				item.Title,
				err,
			)
		}
	}

	return nil
}
