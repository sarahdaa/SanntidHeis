package logic

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/timer"
)

func CyclicCounter(
	hallRequestList 		map[string][][]config.Counter,
	elevator			 	*config.Elevator,
	numberOfPeers 			int,

	) map[string][][]config.Counter {

	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons-1; btn++ {

			if (floor == 0 && btn == int(elevio.BT_HallDown)) ||
				(floor == config.NumFloors-1 && btn == int(elevio.BT_HallUp)) {
				continue
			}

			localState := hallRequestList[elevator.ID][floor][btn]
			

			if localState == config.Initial && numberOfPeers == 1 {
				hallRequestList[elevator.ID][floor][btn] = config.NoOrder
				continue
			}

			if localState == config.Initial {
				for elev := range hallRequestList {
					if elev != elevator.ID && hallRequestList[elev][floor][btn] != config.Initial {
						localState = hallRequestList[elev][floor][btn]
						break
					}
				}
				hallRequestList[elevator.ID][floor][btn] = localState
				continue
			}

			someUnconfirmed := false
			allUnconfirmed := true
			someNoOrder := false
			someConfirmed := false

			
			for elev := range hallRequestList {
				if elev != elevator.ID {
					switch hallRequestList[elev][floor][btn] {
					case config.Unconfirmed:
						someUnconfirmed = true
					case config.Confirmed:
						allUnconfirmed = false
						someConfirmed = true
					case config.NoOrder:
						someNoOrder = true
						allUnconfirmed = false
					}
				}
			}

			switch localState {
			case config.NoOrder:
				if someUnconfirmed && !someConfirmed {
					localState = config.Unconfirmed
				}

			case config.Unconfirmed:
				if someConfirmed {
					localState = config.Confirmed
				} else if allUnconfirmed {
					localState = config.Confirmed
				}

			case config.Confirmed:
				if someNoOrder && !someUnconfirmed {
					localState = config.NoOrder
				}
			}

			hallRequestList[elevator.ID][floor][btn] = localState
		}
	}
	return hallRequestList
}


func HasOrdersAbove(elevator *config.Elevator) bool {
	for floor := elevator.Floor + 1; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if elevator.MyOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func HasOrdersBelow(elevator *config.Elevator) bool {
	for floor := 0; floor < elevator.Floor; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if elevator.MyOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func HasOrdersAt(elevator *config.Elevator) bool {
	for btn := 0; btn < config.NumButtons; btn++ {
		if elevator.MyOrders[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

func ClearRequestsAtFloor(elevator *config.Elevator) {

	floor := elevator.Floor
	currentDir := elevator.CurrDirn
	myOrders := elevator.MyOrders
	hallRequests := elevator.HallRequests


	myOrders[floor][elevio.BT_Cab] = false

	switch currentDir {

	case elevio.MD_Up:
		myOrders[floor][elevio.BT_HallUp] = false
		
		if !HasOrdersAbove(elevator) && hallRequests[floor][elevio.BT_HallUp] != config.Confirmed{
			myOrders[floor][elevio.BT_HallDown] = false
			if floor > 0 {
				hallRequests[floor][elevio.BT_HallDown] = config.NoOrder
			}
		}

		if floor < config.NumFloors-1 {
			hallRequests[floor][elevio.BT_HallUp] = config.NoOrder
		} 

	case elevio.MD_Down:
		myOrders[floor][elevio.BT_HallDown] = false

		if !HasOrdersBelow(elevator) && hallRequests[floor][elevio.BT_HallDown] != config.Confirmed {
			myOrders[floor][elevio.BT_HallUp] = false
			if floor < config.NumFloors-1 {
				hallRequests[floor][elevio.BT_HallUp] = config.NoOrder
			}
		}

		if floor > 0 {
			hallRequests[floor][elevio.BT_HallDown] = config.NoOrder
		}
	
	case elevio.MD_Stop:
		myOrders[floor][elevio.BT_HallUp] = false
		myOrders[floor][elevio.BT_HallDown] = false
		if floor > 0 {
			hallRequests[floor][elevio.BT_HallDown] = config.NoOrder
		}
		if floor < config.NumFloors-1 {
			hallRequests[floor][elevio.BT_HallUp] = config.NoOrder
		}
	}
}



func AddOrder(
	elevator 			*config.Elevator,
	floor 				int,
	btn 				elevio.ButtonType,
	networkChannels 	*config.NetworkChannels,
	timer 				timer.TimerChannels,
	) {

	if btn == elevio.BT_Cab {
		elevator.MyOrders[floor][btn] = true
		networkChannels.UpdatedMyOrders <- elevator.MyOrders

	} else {
		if elevator.HallRequests[floor][btn] != config.Confirmed {
			elevator.HallRequests[floor][btn] = config.Unconfirmed
		}
	}
	networkChannels.StateTX <- *elevator
}


func DeepCopyOrders(original [][]bool) [][]bool {

	copyOrders := make([][]bool, len(original))
	for i := range original {
		copyOrders[i] = make([]bool, len(original[i]))
		copy(copyOrders[i], original[i])
	}
	return copyOrders
}

func DeepCopyRequests(requests [][]config.Counter) [][]config.Counter {

	copyRequests := make([][]config.Counter, len(requests))
	for i := range requests {
		copyRequests[i] = make([]config.Counter, len(requests[i]))
		copy(copyRequests[i], requests[i])
	}
	return copyRequests
}


func DeepCopyElevator(orig config.Elevator) config.Elevator {
	copyElev := orig

	copyElev.MyOrders = make([][]bool, len(orig.MyOrders))
	for i := range orig.MyOrders {
		copyElev.MyOrders[i] = make([]bool, len(orig.MyOrders[i]))
		copy(copyElev.MyOrders[i], orig.MyOrders[i])
	}

	copyElev.HallRequests = make([][]config.Counter, len(orig.HallRequests))
	for i := range orig.HallRequests {
		copyElev.HallRequests[i] = make([]config.Counter, len(orig.HallRequests[i]))
		copy(copyElev.HallRequests[i], orig.HallRequests[i])
	}

	return copyElev
}
