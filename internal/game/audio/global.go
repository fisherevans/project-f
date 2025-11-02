package audio

import "fmt"

var system *System

func init() {
	var err error
	system, err = newSystem()
	if err != nil {
		panic(fmt.Sprintf("failed to start audio system: %v", err))
	}
}

func GetSystem() *System {
	return system
}
