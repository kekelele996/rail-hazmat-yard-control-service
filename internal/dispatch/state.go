package dispatch

import "fmt"

type State string

const (
	StateQueued    State = "queued"
	StateMoving    State = "moving"
	StateHeld      State = "held"
	StateRetrying  State = "retrying"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
)

var transitions = map[State]map[State]bool{StateQueued: {StateMoving: true, StateHeld: true}, StateMoving: {StateCompleted: true, StateFailed: true}, StateHeld: {StateRetrying: true, StateFailed: true}, StateRetrying: {StateMoving: true, StateCompleted: true, StateFailed: true}}

func CanTransition(from, to State) bool { return transitions[from][to] }
func Transition(from, to State) (State, error) {
	if !CanTransition(from, to) {
		return from, fmt.Errorf("invalid dispatch transition %s -> %s", from, to)
	}
	return to, nil
}
func RecoveryPath(from State) ([]State, error) {
	if from != StateHeld {
		return nil, fmt.Errorf("dispatch recovery requires held state, got %s", from)
	}
	return []State{StateRetrying, StateCompleted}, nil
}
func ApplyPath(from State, path []State) (State, error) {
	current := from
	for _, target := range path {
		next, err := Transition(current, target)
		if err != nil {
			return from, err
		}
		current = next
	}
	return current, nil
}
