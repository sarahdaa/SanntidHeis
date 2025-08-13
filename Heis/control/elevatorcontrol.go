package control

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/timer"
	"G19_heis2/Heis/logic"
	"fmt"
)


func ShouldStop(elevator *config.Elevator) bool {
	
	orders := elevator.MyOrders
	currentFloor := elevator.Floor
	currentDir := elevator.CurrDirn

	if elevio.GetFloor() == -1 {
		return false
	}

	if orders[currentFloor][elevio.BT_Cab] {
		return true
	}

	if currentDir == elevio.MD_Up && orders[currentFloor][elevio.BT_HallUp] {
		return true
	}
	if currentDir == elevio.MD_Down && orders[currentFloor][elevio.BT_HallDown] {
		return true
	}

	if (currentDir == elevio.MD_Up && !logic.HasOrdersAbove(elevator)) ||
		(currentDir == elevio.MD_Down && !logic.HasOrdersBelow(elevator)) {
		return true
	}

	if currentDir == elevio.MD_Stop && logic.HasOrdersAt(elevator) {
		return true
	}

	return false
}


func StopElevator(
	elevator 				*config.Elevator,
	timer 					timer.TimerChannels,
	LightCh 				*config.LightChannels,
	){
		elevio.SetMotorDirection(elevio.MD_Stop)
		logic.ClearRequestsAtFloor(elevator)
		LightCh.MyOrdersForLight <- elevator.MyOrders

		doorClosed := make(chan bool)

		go OpenDoor(elevator, timer, doorClosed)

		go func() {
			<-doorClosed 

			elevator.CurrDirn, elevator.Behaviour = ChooseDirection(elevator)
			elevio.SetMotorDirection(elevator.CurrDirn)
		}()
}


func UpdateHallButtonLights(requests map[string][][2]bool) {
	
	numFloors := config.NumFloors

	for floor := 0; floor < numFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			shouldBeLit := false

			for _, peerRequests := range requests {
				if len(peerRequests) <= floor || len(peerRequests[floor]) <= btn {
					continue 
				}
				if peerRequests[floor][btn]{
					shouldBeLit = true
					break
				}
			}

			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, shouldBeLit)
		}
	}
}


func UpdateMyOrdersLights(myOrders [][]bool) {

	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if myOrders[floor][btn] {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func LightController(lightChannels *config.LightChannels) {
	for {
		select {
		case requests := <-lightChannels.HallRequestsForLight:
			UpdateHallButtonLights(requests)

		case myOrders := <-lightChannels.MyOrdersForLight:
			UpdateMyOrdersLights(myOrders)

		case newFloor := <-lightChannels.FloorIndicatorForLight:
			elevio.SetFloorIndicator(newFloor)
		}
	}
}

func OpenDoor(
	elevator 				*config.Elevator,
	timer 					timer.TimerChannels,
	doorClosed 				chan bool,
	) {

	elevator.Behaviour = config.DOOR_OPEN
	elevio.SetDoorOpenLamp(true)

	timer.OpenDoorChan <- true

	for {
		select {
		case <-timer.CloseDoorChan:
			if !config.Notimer { 
				elevio.SetDoorOpenLamp(false)
				doorClosed <- true 
				return
			} else {
				fmt.Println("Obstruction still present, keeping door open")
			}
		}
	}
}


func ChooseDirection(elevator *config.Elevator) (elevio.MotorDirection, config.ElevatorBehaviour) {
	orders := elevator.MyOrders

	if elevator.CurrDirn == elevio.MD_Stop {
		for floor := 0; floor < config.NumFloors; floor++ {
			for btn := 0; btn < config.NumButtons; btn++ {
				if orders[floor][btn] {
					if floor > elevator.Floor {
						return elevio.MD_Up, config.MOVING
					} else if floor < elevator.Floor {
						return elevio.MD_Down, config.MOVING
					}
				}
			}
		}
	}

	if elevator.CurrDirn == elevio.MD_Up {
		if logic.HasOrdersAbove(elevator) {
			return elevio.MD_Up, config.MOVING
		}
		if logic.HasOrdersBelow(elevator) {
			return elevio.MD_Down, config.MOVING
		}
	}

	if elevator.CurrDirn == elevio.MD_Down {
		if logic.HasOrdersBelow(elevator) {
			return elevio.MD_Down, config.MOVING
		}
		if logic.HasOrdersAbove(elevator) {
			return elevio.MD_Up, config.MOVING
		}
	}

	if logic.HasOrdersAbove(elevator) {
		return elevio.MD_Up, config.MOVING
	}
	if logic.HasOrdersBelow(elevator) {
		return elevio.MD_Down, config.MOVING
	}

	return elevio.MD_Stop, config.IDLE
}
