package core

type State []byte

type SaveStateFunc func(id string, state State) error

type LoadStateFunc func(id string) State
