package timer

import (
	"G19_heis2/Heis/config"
	"time"
)

const doorOpenDuration = 3 * time.Second 
const motorPowerTimeout = 10 * time.Second

var motorTimerStarted bool

type TimerChannels struct {
	OpenDoorChan  			chan bool
	CloseDoorChan 			chan bool
	StartMotorTimer			chan bool
	StopMotorTimer 			chan bool
}

func NewTimerChannels() *TimerChannels {
	return &TimerChannels		{
		OpenDoorChan:  			make(chan bool),
		CloseDoorChan: 			make(chan bool),
		StartMotorTimer: 		make(chan bool),
		StopMotorTimer: 		make(chan bool),
	}
}


func Timer(
	tc 					*TimerChannels,
	drv_obstr 			chan bool,
	elevator 			*config.Elevator,
	) {
	
	var startDoor bool
	var obstruction bool
	var motorRunning bool
	

	doorTimer := time.NewTimer(time.Hour)
	doorTimer.Stop()

	motorTimer := time.NewTimer(time.Hour)
	motorTimer.Stop()

	for {
		select {
		case startDoor = <-tc.OpenDoorChan:
			if !doorTimer.Stop() {
				select {
				case <-doorTimer.C:
				default:
				}
			}
			doorTimer.Reset(doorOpenDuration)

		case obstruction = <-drv_obstr:
			if obstruction {
				doorTimer.Stop() 
			} else {
				doorTimer.Reset(doorOpenDuration) 
			}

		case <-doorTimer.C:
			if startDoor && !obstruction {
				startDoor = false
				tc.CloseDoorChan <- true
			}

		case <- tc.StartMotorTimer:
			if motorTimerStarted{
				break
			}
			if !motorTimer.Stop() {
				select {
				case <-motorTimer.C:
				default:
				}
			}
			motorTimer.Reset(motorPowerTimeout)
			motorRunning = true
			motorTimerStarted = true

		case <-tc.StopMotorTimer:
			if !motorTimer.Stop() {
				select {
				case <-motorTimer.C:
				default:
				}
			}
			elevator.MotorPower = true
			motorTimerStarted = false
		

		case <- motorTimer.C:
			if motorRunning {
				elevator.MotorPower = false
				motorRunning = false
			}
		}
	}
}

func StartingMotorTimer(elevator *config.Elevator, timerChannels TimerChannels) {
	for {
		if elevator.Behaviour != config.IDLE && !motorTimerStarted {
			select {
			case timerChannels.StartMotorTimer <- true:
			default:
			}
		}
		if elevator.Behaviour == config.IDLE {
			timerChannels.StopMotorTimer <- true
		}
		time.Sleep(100 * time.Millisecond)
	}
}
