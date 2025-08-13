package main

import (
	"G19_heis2/Heis/fsm"
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/control"
	"G19_heis2/Heis/network"
	"G19_heis2/Heis/timer"
	"fmt"
)


func main() {

	var (
		networkChannels = &config.NetworkChannels{
			UpdatedMyOrders: make(chan [][]bool, config.Buffsize),
		}
		lightChannels = &config.LightChannels{
			HallRequestsForLight:    make(chan map[string][][2]bool, config.Buffsize),
			MyOrdersForLight:        make(chan [][]bool, config.Buffsize),
			FloorIndicatorForLight:  make(chan int, config.Buffsize),
		}
		drvChannels = config.DrvChannels{
			Buttons: make(chan elevio.ButtonEvent),
			Floors:  make(chan int),
			Obstr:   make(chan bool),
			Stop:    make(chan bool),
		}
	)

	elevio.Init("localhost:15657", config.NumFloors) 
	id := config.InitID()
	elevator := config.InitElev(id)

	timerChannels := timer.NewTimerChannels()
	

	go elevio.PollButtons(drvChannels.Buttons)
	go elevio.PollFloorSensor(drvChannels.Floors)
	go elevio.PollObstructionSwitch(drvChannels.Obstr)
	go elevio.PollStopButton(drvChannels.Stop)

	fmt.Println("Elevator system initialized...")

	go network.Networking(&elevator, networkChannels, lightChannels)
	go control.LightController(lightChannels)
	go fsm.Fsm(&elevator, drvChannels, networkChannels, *timerChannels, lightChannels)
	go timer.Timer(timerChannels, drvChannels.Obstr, &elevator)
	go timer.StartingMotorTimer(&elevator, *timerChannels)
	select {}
}
