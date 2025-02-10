package main

import (
	"fmt"
	//"time"
	"G19_heis2/Heis/driver/elevio"
	"G19_heis2/Heis/config"
	//"G19_heis2/Heis/logic"
	"G19_heis2/Heis/FSM"
	"G19_heis2/Heis/failuredetection"
)

func main() {
	numFloors := 4

	// Initialize the elevator system
	elevio.Init("localhost:15657", numFloors)
	id := config.InitID()
	elevator := config.InitElev(id)

	//var currentFloor int = 0
	//var currentDir elevio.MotorDirection = elevio.MD_Stop

	txHeartbeat := make(chan failuredetection.HeartBeat)
	rxHeartbeat := make(chan failuredetection.HeartBeat)

	failuredetection.StartHeartBeat(&elevator, txHeartbeat, rxHeartbeat)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	fmt.Println("Elevator system initialized...")

	FSM.Fsm(&elevator, drv_buttons, drv_obstr, drv_stop, drv_floors, numFloors)
}
