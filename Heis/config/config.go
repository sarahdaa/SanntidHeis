package config

import (
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/network/localip"
	"G19_heis2/Heis/network/peers"
	"flag"
	"fmt"
	"os"
)

var Notimer bool

const (
	Buffsize = 100
)

const (
	NumButtons = 3
	NumFloors  = 4
)

type ElevatorBehaviour int

const (
	IDLE ElevatorBehaviour = iota
	MOVING
	DOOR_OPEN
	STOPPED
)

type Counter int 

const (
	Initial Counter = iota
	NoOrder 
	Unconfirmed 
	Confirmed 
)


type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    	[][2]bool                   `json:"hallRequests"`
    State          	map[string]HRAElevState     `json:"states"`
}

type Elevator struct {
	ID 						string
	Floor 					int
	CurrDirn 				elevio.MotorDirection
	HallRequests 			[][]Counter 
	MyOrders 				[][]bool
	Behaviour 				ElevatorBehaviour
	MotorPower 				bool
}

type NetworkChannels struct {
	StateRX 					chan Elevator
	StateTX 					chan Elevator 
	PeerUpdate 					chan peers.PeerUpdate
	UpdatedMyOrders 			chan [][]bool
}

type LightChannels struct {
	HallRequestsForLight 		chan map[string][][2]bool
	MyOrdersForLight 			chan [][]bool
	FloorIndicatorForLight 		chan int
}

type DrvChannels struct {
	Buttons  			chan elevio.ButtonEvent
	Floors  			chan int
	Obstr 				chan bool
	Stop 				chan bool
}



func InitElev(ID string) Elevator {

	requests := make([][]Counter, NumFloors)
	orders := make([][]bool, NumFloors)

	for i := range requests {
		requests[i] = make([]Counter, NumButtons-1)
		for j := range requests[i]{
			requests[i][j] = Initial
		}
	}

	for i := range orders {
		orders[i] = make([]bool, NumButtons)
	}


	for floor := elevio.GetFloor(); floor == -1; floor = elevio.GetFloor(){
		elevio.SetMotorDirection(elevio.MD_Down)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)

	return Elevator{
		ID: 					ID,
		Floor: 					elevio.GetFloor(),
		CurrDirn: 				elevio.MD_Stop,
		HallRequests: 			requests,
		MyOrders: 				orders,
		Behaviour: 				IDLE,
		MotorPower: 			true,
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

