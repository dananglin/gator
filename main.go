package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"codeflow.dananglin.me.uk/apollo/gator/internal/config"
	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	"codeflow.dananglin.me.uk/apollo/gator/internal/executors"
	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
	_ "github.com/lib/pq"
)

var (
	binaryVersion string
	buildTime     string
	goVersion     string
	gitCommit     string
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v.\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("unable to load the configuration: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DBConfig.URL)
	if err != nil {
		return fmt.Errorf("unable to open a connection to the database: %w", err)
	}

	s := state.State{
		DB:     database.New(db),
		Config: &cfg,
	}

	executorMap := executors.ExecutorMap{
		Map: make(map[string]executors.ExecutorFunc),
	}

	executorMap.Register("login", executors.Login)
	executorMap.Register("register", executors.Register)
	executorMap.Register("reset", executors.Reset)
	executorMap.Register("users", executors.Users)
	executorMap.Register("aggregate", executors.Aggregate)
	executorMap.Register("addfeed", executors.MiddlewareLoggedIn(executors.AddFeed))
	executorMap.Register("feeds", executors.Feeds)
	executorMap.Register("follow", executors.MiddlewareLoggedIn(executors.Follow))
	executorMap.Register("unfollow", executors.MiddlewareLoggedIn(executors.Unfollow))
	executorMap.Register("following", executors.MiddlewareLoggedIn(executors.Following))
	executorMap.Register("browse", executors.MiddlewareLoggedIn(executors.Browse))

	executor, err := parseArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("unable to parse the command: %w", err)
	}

	return executorMap.Run(&s, executor)
}

func parseArgs(args []string) (executors.Executor, error) {
	if len(args) == 0 {
		return executors.Executor{}, errors.New("no arguments given")
	}

	if len(args) == 1 {
		return executors.Executor{
			Name: args[0],
			Args: make([]string, 0),
		}, nil
	}

	return executors.Executor{
		Name: args[0],
		Args: args[1:],
	}, nil
}
