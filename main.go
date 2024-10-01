package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"codeflow.dananglin.me.uk/apollo/gator/internal/config"
	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

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

	s := state{
		db:     database.New(db),
		config: &cfg,
	}

	cmds := commands{
		commandMap: make(map[string]commandFunc),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	cmd, err := parseArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("unable to parse the command: %w", err)
	}

	return cmds.run(&s, cmd)
}

func parseArgs(args []string) (command, error) {
	if len(args) == 0 {
		return command{}, errors.New("no arguments given")
	}

	if len(args) == 1 {
		return command{
			name: args[0],
			args: make([]string, 0),
		}, nil
	}

	return command{
		name: args[0],
		args: args[1:],
	}, nil
}
