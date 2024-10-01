package executors

import (
	"fmt"

	"codeflow.dananglin.me.uk/apollo/gator/internal/state"
)

type ExecutorMap struct {
	Map map[string]ExecutorFunc
}

type ExecutorFunc func(*state.State, Executor) error

type Executor struct {
	Name string
	Args []string
}

func (e *ExecutorMap) Register(name string, f ExecutorFunc) {
	e.Map[name] = f
}

func (e *ExecutorMap) Run(s *state.State, exe Executor) error {
	runFunc, ok := e.Map[exe.Name]
	if !ok {
		return fmt.Errorf("unrecognised command: %s", exe.Name)
	}

	return runFunc(s, exe)
}
