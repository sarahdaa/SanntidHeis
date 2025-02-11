package config

//"fmt"
//"G19_heis2/Heis/driver/elevio"
import (
	"G19_heis2/Heis/driver/elevio"
	"G19_heis2/Heis/network/localip"
	"flag"
	"os"
	"fmt"
)

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

type Elevator struct {
	ID string
	Floor int
	CurrDirn elevio.MotorDirection
	Requests [][]bool
	State ElevatorState 
	IsOnline bool
}


func InitElev(ID string) Elevator {
	requests := make([][]bool, NumFloors)

	for i := range requests {
		requests[i] = make([]bool, NumButtons)
	}

	for floor := elevio.GetFloor(); floor == -1; floor = elevio.GetFloor(){
		elevio.SetMotorDirection(elevio.MD_Down)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)

	return Elevator{
		ID: ID,
		Floor: elevio.GetFloor(),
		CurrDirn: elevio.MD_Stop,
		Requests: requests,
		State: IDLE,
		IsOnline: true,
	}
}

func InitID() string{
	idPtr := flag.String("Id","","Id of this elevator")
	flag.Parse()

	if *idPtr != ""{
		return *idPtr
	}

	localIP,err:= localip.LocalIP()
	if err!= nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not retrieve local IP: %v/n", err)
		localIP = "Unknown"

	}
	return fmt.Sprintf("%s-%d", localIP, os.Getpid())
}

type HallRequestAssignment struct {
	ID string 
	UpRequests []bool
	DownRequests []bool
}
type AssignmentResults struct{
	Assignments []HallRequestAssignment
}

type elevatorchannels struct {
	Drv_buttons chan elevio.ButtonEvent
	Drv_floors chan int 
	Drv_obstruction  chan bool
	AssignHallOrders chan elevio.ButtonEvent
	HallOrders chan *AssignmentResults

}

