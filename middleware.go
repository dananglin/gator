package main

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) commandFunc {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.config.CurrentUsername)
		if err != nil {
			return fmt.Errorf("unable to get the user from the database: %w", err)
		}

		return handler(s, cmd, user)
	}
}
