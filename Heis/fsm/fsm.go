package fsm

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/control"
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/logic"
	"G19_heis2/Heis/timer"
	"fmt"
	"time"
)

func Fsm(
	elevator 					*config.Elevator,
	drvChannels 				config.DrvChannels,
	networkChannels 			*config.NetworkChannels,
	timer 						timer.TimerChannels,
	lightChannels 				*config.LightChannels) {

	for {

		select {

		case btnPress := <-drvChannels.Buttons:
			fmt.Printf("Button pressed: %+v\n", btnPress)
			logic.AddOrder(elevator, btnPress.Floor, btnPress.Button, networkChannels, timer)

			
		case NewMyOrder := <- networkChannels.UpdatedMyOrders:
			elevator.MyOrders = NewMyOrder
			lightChannels.MyOrdersForLight <- elevator.MyOrders

			if elevator.Behaviour != config.DOOR_OPEN{
				newDirection, newBehaviour := control.ChooseDirection(elevator)

				if newDirection != elevator.CurrDirn {
					elevator.CurrDirn = newDirection
					elevator.Behaviour = newBehaviour
					
					elevio.SetMotorDirection(elevator.CurrDirn)
				}
				if control.ShouldStop(elevator) {
					control.StopElevator(elevator, timer, lightChannels)
				} 
			} 
			

		case newFloor := <-drvChannels.Floors:
			elevator.Floor = newFloor
			timer.StopMotorTimer <- true
			
			lightChannels.FloorIndicatorForLight <- newFloor
			lightChannels.MyOrdersForLight <- elevator.MyOrders

			if control.ShouldStop(elevator) {
				control.StopElevator(elevator, timer, lightChannels)
			} 

		case <-drvChannels.Stop:
			fmt.Println("Emergency stop button pressed!")
			elevio.SetMotorDirection(elevio.MD_Stop)

			elevio.SetStopLamp(true)
			time.Sleep(3 * time.Second)
			elevio.SetStopLamp(false)

			if elevator.Behaviour != config.DOOR_OPEN{
				elevator.CurrDirn, _ = control.ChooseDirection(elevator)
				elevio.SetMotorDirection(elevator.CurrDirn)
			}
			

		default:
			time.Sleep(time.Millisecond * 50)

		}
	}
}
