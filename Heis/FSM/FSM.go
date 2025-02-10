package FSM

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/driver/elevio"
	"G19_heis2/Heis/logic"
	"fmt"
	"time"
)

func Fsm(elevator *config.Elevator, drv_buttons chan elevio.ButtonEvent, drv_obstr chan bool, drv_stop chan bool, drv_floors chan int, numFloors int) {
	for {
		select {

		// Handling button presses
		case btnPress := <-drv_buttons:
			fmt.Printf("Button pressed: %+v\n", btnPress)
			logic.AddOrder(btnPress.Floor, btnPress.Button)

			if elevator.CurrDirn == elevio.MD_Stop {
				elevator.CurrDirn = logic.ChooseDirection(elevator.Floor, elevator.CurrDirn, config.Orders)
				elevio.SetMotorDirection(elevator.CurrDirn)
			}

		// Handling floor sensor updates
		case newFloor := <-drv_floors:
			fmt.Printf("Arrived at floor: %d\n", newFloor)
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)

			logic.ControlElevator(elevator.Floor, &elevator.CurrDirn, &config.Orders)

		// Handling obstruction events
		case obstruction := <-drv_obstr:
			fmt.Printf("Obstruction detected: %t\n", obstruction)
			if obstruction {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevator.CurrDirn = logic.ChooseDirection(elevator.Floor, elevator.CurrDirn, config.Orders)
				elevio.SetMotorDirection(elevator.CurrDirn)
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

			elevator.CurrDirn = logic.ChooseDirection(elevator.Floor, elevator.CurrDirn, config.Orders)
			elevio.SetMotorDirection(elevator.CurrDirn)

		}
	}
}
