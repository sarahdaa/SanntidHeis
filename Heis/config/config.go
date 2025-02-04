package config

//"fmt"
//"G19_heis2/Heis/driver/elevio"

const (
	NumButtons = 3
	NumFloors  = 4
)

var Orders [NumFloors][NumButtons]bool

type ElevatorState int

const (
	IDLE ElevatorState = iota
	MOVING
	DOOR_OPEN
	STOPPED
)
