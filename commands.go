package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type commands struct {
	commandMap map[string]commandFunc
}

type commandFunc func(*state, command) error

type command struct {
	name string
	args []string
}

func (c *commands) register(name string, f commandFunc) {
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	runFunc, ok := c.commandMap[cmd.name]
	if !ok {
		return fmt.Errorf("unrecognised command: %s", cmd.name)
	}

	return runFunc(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(cmd.args))
	}

	username := cmd.args[0]

	user, err := s.db.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("unable to get the user from the database: %w", err)
	}

	if err := s.config.SetUser(user.Name); err != nil {
		return fmt.Errorf("login error: %w", err)
	}

	fmt.Printf("The current user is set to %q.\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(cmd.args))
	}

	name := cmd.args[0]

	timestamp := time.Now()

	args := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), args)
	if err != nil {
		if uniqueViolation(err) {
			return errors.New("this user is already registered")
		}

		return fmt.Errorf("unable to register the user: %w", err)
	}

	if err := s.config.SetUser(name); err != nil {
		return fmt.Errorf("unable to update the configuration: %w", err)
	}

	fmt.Printf("Successfully registered %s.\n", user.Name)
	fmt.Println("DEBUG:", user)

	return nil
}

func handlerReset(s *state, _ command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		fmt.Errorf("unable to delete the users from the database: %w", err)
	}

	fmt.Println("Successfully removed all users from the database.")

	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		fmt.Errorf("unable to get the users from the database: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("There are no registered users.")

		return nil
	}

	fmt.Printf("Registered users:\n\n")

	for _, user := range users {
		if user.Name == s.config.CurrentUsername {
			fmt.Printf("- %s (current)\n", user.Name)
		} else {
			fmt.Printf("- %s\n", user.Name)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(cmd.args))
	}

	intervalArg := cmd.args[0]

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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("unexpected number of arguments: want 2, got %d", len(cmd.args))
	}

	name, url := cmd.args[0], cmd.args[1]

	timestamp := time.Now()

	createdFeedArgs := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), createdFeedArgs)
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

	followRecord, err := s.db.CreateFeedFollow(context.Background(), createFeedFollowArgs)
	if err != nil {
		return fmt.Errorf("unable to create the feed follow record in the database: %w", err)
	}

	fmt.Printf("You are now following the feed %q.\n", followRecord.FeedName)
	fmt.Println("DEBUG:", followRecord)

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get the feeds from the database: %w", err)
	}

	fmt.Printf("Feeds:\n\n")

	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 2, got %d", len(cmd.args))
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
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

	followRecord, err := s.db.CreateFeedFollow(context.Background(), args)
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

func handlerFollowing(s *state, _ command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 2, got %d", len(cmd.args))
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("unable to get the feed data from the database: %w", err)
	}

	args := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	if err := s.db.DeleteFeedFollow(context.Background(), args); err != nil {
		return fmt.Errorf("unable to delete the feed follow record from the database: %w", err)
	}

	fmt.Printf("You have successfully unfollowed %q.\n", feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return fmt.Errorf("unexpected number of arguments: want 0 or 1, got %d", len(cmd.args))
	}

	var err error

	limit := 2

	if len(cmd.args) == 1 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("unable to convert %s to a number: %w", cmd.args[0], err)
		}
	}

	args := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.db.GetPostsForUser(context.Background(), args)
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

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
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

	if err := s.db.MarkFeedFetched(context.Background(), markFeedFetchedArgs); err != nil {
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

		_, err = s.db.CreatePost(context.Background(), args)
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

func uniqueViolation(err error) bool {
	var pqError *pq.Error

	if errors.As(err, &pqError) {
		if pqError.Code.Name() == "unique_violation" {
			return true
		}
	}

	return false
}
