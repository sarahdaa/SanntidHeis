package FSM

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/driver/elevio"
	"G19_heis2/Heis/logic"
	"fmt"
	"time"
)

func Fsm(currentDir elevio.MotorDirection, currentFloor int, drv_buttons chan elevio.ButtonEvent, drv_obstr chan bool, drv_stop chan bool, drv_floors chan int, numFloors int) {
	for {
		select {

		// Handling button presses
		case btnPress := <-drv_buttons:
			fmt.Printf("Button pressed: %+v\n", btnPress)
			logic.AddOrder(btnPress.Floor, btnPress.Button)

			if currentDir == elevio.MD_Stop {
				currentDir = logic.ChooseDirection(currentFloor, currentDir, config.Orders)
				elevio.SetMotorDirection(currentDir)
			}

		// Handling floor sensor updates
		case newFloor := <-drv_floors:
			fmt.Printf("Arrived at floor: %d\n", newFloor)
			currentFloor = newFloor
			elevio.SetFloorIndicator(currentFloor)

			logic.ControlElevator(currentFloor, &currentDir, &config.Orders)

		// Handling obstruction events
		case obstruction := <-drv_obstr:
			fmt.Printf("Obstruction detected: %t\n", obstruction)
			if obstruction {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				currentDir = logic.ChooseDirection(currentFloor, currentDir, config.Orders)
				elevio.SetMotorDirection(currentDir)
			}

			// Handling stop button press
		case <-drv_stop:
			fmt.Println("Emergency stop button pressed!")
			elevio.SetMotorDirection(elevio.MD_Stop)

			// Clear all orders and lights
			for f := 0; f < numFloors; f++ {
				for b := 0; b < 3; b++ {
					logic.RemoveOrder(f, elevio.ButtonType(b))
				}
			}

			elevio.SetStopLamp(true)
			time.Sleep(3 * time.Second)
			elevio.SetStopLamp(false)

			currentDir = logic.ChooseDirection(currentFloor, currentDir, config.Orders)
			elevio.SetMotorDirection(currentDir)

		}
	}
}
