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

func Register(s *state.State, exe Executor) error {
	if len(exe.Args) != 1 {
		return fmt.Errorf("unexpected number of arguments: want 1, got %d", len(exe.Args))
	}

	name := exe.Args[0]

	timestamp := time.Now()

	args := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
		Name:      name,
	}

	user, err := s.DB.CreateUser(context.Background(), args)
	if err != nil {
		if uniqueViolation(err) {
			return errors.New("this user is already registered")
		}

		return fmt.Errorf("unable to register the user: %w", err)
	}

	if err := s.Config.SetUser(name); err != nil {
		return fmt.Errorf("unable to update the configuration: %w", err)
	}

	fmt.Printf("Successfully registered %s.\n", user.Name)
	fmt.Println("DEBUG:", user)

	return nil
}
