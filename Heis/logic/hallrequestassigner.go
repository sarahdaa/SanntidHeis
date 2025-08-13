package logic

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
)


func HallRequestAssigner(
	hallRequests [][2]bool,
	states map[string]config.HRAElevState,

	) (map[string][][2]bool, error) {

	hraExecutable := "hall_request_assigner"
	
	input := config.HRAInput{
		HallRequests: hallRequests,
		State:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {

		return nil, fmt.Errorf("json.Marshal error: %v", err)
	}

	ret, err := exec.Command("Heis/logic/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {

		return nil, fmt.Errorf("exec.Command error: %v, output: %s", err, string(ret))
	}

	var output map[string][][2]bool
	if err = json.Unmarshal(ret, &output); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: ", err)
	}

	return output, nil

}


func ElevatorToHRAELEV(elevatormap map[string]config.Elevator) map[string]config.HRAElevState {

	hraElevMap := make(map[string]config.HRAElevState)

	localCopy := make(map[string]config.Elevator)
	for key, value := range elevatormap {
		localCopy[key] = value
	}

	for key, value := range localCopy {
		cabRequests := make([]bool, config.NumFloors)
		for floor := 0; floor < config.NumFloors; floor++ {
			if value.MyOrders[floor][elevio.BT_Cab] {
				cabRequests[floor] = true
			} else {
				cabRequests[floor] = false
			}
		}

		behavior := ""
		switch value.Behaviour {
		case config.IDLE:
			behavior = "idle"
		case config.MOVING:
			behavior = "moving"
		case config.DOOR_OPEN:
			behavior = "doorOpen"
		case config.STOPPED:
			behavior = "stopped"
		default:
			behavior = "unknown"
		}

		direction := ""
		switch value.CurrDirn {
		case elevio.MD_Up:
			direction = "up"
		case elevio.MD_Down:
			direction = "down"
		case elevio.MD_Stop:
			direction = "stop"
		default:
			direction = "unknown"
		}

		hraElevMap[key] = config.HRAElevState{
			Behavior:    behavior,
			Floor:       value.Floor,
			Direction:   direction,
			CabRequests: cabRequests,
		}
	}

	return hraElevMap
}


func CreateHRAInputRequests(elevator [][]config.Counter) [][2]bool {
	hallRequests := make([][2]bool, config.NumFloors)

	for floor := 0; floor < config.NumFloors; floor++ {
		if elevator[floor][elevio.BT_HallUp] == config.Confirmed {
			hallRequests[floor][elevio.BT_HallUp] = true
		}
		if elevator[floor][elevio.BT_HallDown] == config.Confirmed {
			hallRequests[floor][elevio.BT_HallDown] = true
		}
	}
	return hallRequests
}

func UpdateElevatorHallRequests(
	elevator 						*config.Elevator,
	HRAInputRequests 				[][2]bool,
	HRAElevatorMap 					map[string]config.HRAElevState,
	networkChannels 						*config.NetworkChannels,
	lightChannels 					*config.LightChannels,
	) {

	MyOrdersCopy := DeepCopyOrders(elevator.MyOrders)
	

	assignedHallRequests, err := HallRequestAssigner(HRAInputRequests, HRAElevatorMap)
	lightChannels.HallRequestsForLight <- assignedHallRequests
	

	if err != nil {
		fmt.Printf("Feil ved kall til HallRequestAssigner: %v\n", err)
		return
	}

	if _, exists := assignedHallRequests[elevator.ID]; !exists {
		return
	}

	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			MyOrdersCopy[floor][btn] = assignedHallRequests[elevator.ID][floor][btn]
		}
		MyOrdersCopy[floor][elevio.BT_Cab] = elevator.MyOrders[floor][elevio.BT_Cab]
	}
	if !reflect.DeepEqual(MyOrdersCopy, elevator.MyOrders) {
		networkChannels.UpdatedMyOrders <- MyOrdersCopy
	}

}
