package logic

import (
	"G19_heis2/Heis/driver/elevio"
	"time"
)

func ControlElevator(currentFloor int, currentDir *elevio.MotorDirection, orders *[4][3]bool) {
	if ShouldStop(currentFloor, *currentDir, *orders) {
		elevio.SetMotorDirection(elevio.MD_Stop)
		ClearRequestsAtFloor(currentFloor, *currentDir, orders)
		UpdateButtonLights(*orders)
		elevio.SetDoorOpenLamp(true)
		time.Sleep(3 * time.Second)
		elevio.SetDoorOpenLamp(false)
		*currentDir = ChooseDirection(currentFloor, *currentDir, *orders)
		elevio.SetMotorDirection(*currentDir)
	}
}

func UpdateButtonLights(orders [4][3]bool) {
	for floor := 0; floor < len(orders); floor++ {
		for btn := 0; btn < 3; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, orders[floor][btn])
		}
	}
}
