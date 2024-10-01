package state

import (
	"codeflow.dananglin.me.uk/apollo/gator/internal/config"
	"codeflow.dananglin.me.uk/apollo/gator/internal/database"
)

type State struct {
	DB     *database.Queries
	Config *config.Config
}
