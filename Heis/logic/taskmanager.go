package logic

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/driver/elevio"
)

func ChooseDirection(currentFloor int, currentDir elevio.MotorDirection, orders [4][3]bool) elevio.MotorDirection {
	if currentDir == elevio.MD_Up {
		if hasOrdersAbove(currentFloor, orders) {
			return elevio.MD_Up
		}
	} else if currentDir == elevio.MD_Down {
		if hasOrdersBelow(currentFloor, orders) {
			return elevio.MD_Down
		}
	}

	if hasOrdersAbove(currentFloor, orders) {
		return elevio.MD_Up
	}
	if hasOrdersBelow(currentFloor, orders) {
		return elevio.MD_Down
	}

	return elevio.MD_Stop
}

func hasOrdersAbove(floor int, orders [4][3]bool) bool {
	for f := floor + 1; f < len(orders); f++ {
		for btn := 0; btn < 3; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasOrdersBelow(floor int, orders [4][3]bool) bool {
	for f := 0; f < floor; f++ {
		for btn := 0; btn < 3; btn++ {
			if orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasOrdersAt(floor int, orders [4][3]bool) bool {
	for btn := 0; btn < 3; btn++ {
		if orders[floor][btn] {
			return true
		}
	}
	return false
}

func ShouldStop(currentFloor int, currentDir elevio.MotorDirection, orders [4][3]bool) bool {
	if orders[currentFloor][elevio.BT_Cab] {
		return true
	}
	if currentDir == elevio.MD_Up && orders[currentFloor][elevio.BT_HallUp] {
		return true
	}
	if currentDir == elevio.MD_Down && orders[currentFloor][elevio.BT_HallDown] {
		return true
	}
	if (currentDir == elevio.MD_Up && !hasOrdersAbove(currentFloor, orders)) ||
		(currentDir == elevio.MD_Down && !hasOrdersBelow(currentFloor, orders)) {
		return true
	}
	return false
}

func ClearRequestsAtFloor(floor int, currentDir elevio.MotorDirection, orders *[4][3]bool) {
	orders[floor][elevio.BT_Cab] = false
	if currentDir == elevio.MD_Up {
		orders[floor][elevio.BT_HallUp] = false
		if !hasOrdersAbove(floor, *orders) {
			orders[floor][elevio.BT_HallDown] = false
		}
	} else if currentDir == elevio.MD_Down {
		orders[floor][elevio.BT_HallDown] = false
		if !hasOrdersBelow(floor, *orders) {
			orders[floor][elevio.BT_HallUp] = false
		}
	} else {
		orders[floor][elevio.BT_HallUp] = false
		orders[floor][elevio.BT_HallDown] = false
	}
}

func AddOrder(floor int, btn elevio.ButtonType) {
	config.Orders[floor][btn] = true
	UpdateButtonLights(config.Orders)
}

func RemoveOrder(floor int, btn elevio.ButtonType) {
	config.Orders[floor][btn] = false
	elevio.SetButtonLamp(btn, floor, false)
}
