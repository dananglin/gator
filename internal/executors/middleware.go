package executors

import (
	"context"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

func MiddlewareLoggedIn(handler func(s *state.State, exe Executor, user database.User) error) ExecutorFunc {
	return func(s *state.State, exe Executor) error {
		user, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUsername)
		if err != nil {
			return fmt.Errorf("unable to get the user from the database: %w", err)
		}

		return handler(s, exe, user)
	}
}
