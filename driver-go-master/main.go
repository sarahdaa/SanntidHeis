package main

import (
    "fmt"
    "time"
    "Driver-go/elevio"
)

func main() {
    numFloors := 4

    // Initialize the elevator system
    elevio.Init("localhost:15657", numFloors)

    var currentFloor int = 0
    var currentDir elevio.MotorDirection = elevio.MD_Stop

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)

    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    fmt.Println("Elevator system initialized...")

    for {
        select {

        // Handling button presses
        case btnPress := <-drv_buttons:
            fmt.Printf("Button pressed: %+v\n", btnPress)
            elevio.AddOrder(btnPress.Floor, btnPress.Button)

            if currentDir == elevio.MD_Stop {
                currentDir = elevio.ChooseDirection(currentFloor, currentDir, elevio.Orders)
                elevio.SetMotorDirection(currentDir)
            }

        // Handling floor sensor updates
        case newFloor := <-drv_floors:
            fmt.Printf("Arrived at floor: %d\n", newFloor)
            currentFloor = newFloor
            elevio.SetFloorIndicator(currentFloor)

            elevio.ControlElevator(currentFloor, &currentDir, &elevio.Orders)

        // Handling obstruction events
        case obstruction := <-drv_obstr:
            fmt.Printf("Obstruction detected: %t\n", obstruction)
            if obstruction {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                currentDir = elevio.ChooseDirection(currentFloor, currentDir, elevio.Orders)
                elevio.SetMotorDirection(currentDir)
            }

        // Handling stop button press
    case <-drv_stop:
        fmt.Println("Emergency stop button pressed!")
        elevio.SetMotorDirection(elevio.MD_Stop)
    
        // Clear all orders and lights
        for f := 0; f < numFloors; f++ {
            for b := 0; b < 3; b++ {
                elevio.RemoveOrder(f, elevio.ButtonType(b))
            }
        }
    
        elevio.SetStopLamp(true)
        time.Sleep(3 * time.Second)
        elevio.SetStopLamp(false)
    
        currentDir = elevio.ChooseDirection(currentFloor, currentDir, elevio.Orders)
        elevio.SetMotorDirection(currentDir)
    
        }
    }
}
