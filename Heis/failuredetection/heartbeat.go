package failuredetection

import (
	//"fmt"
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/network/bcast"
	"fmt"
	"time"
)

type HeartBeat struct {
	ElevatorID string
	Timestamp time.Time
}

func StartHeartBeat(Elevator *config.Elevator, tx chan HeartBeat, rx chan HeartBeat){
	peerLastSeen := make(map[string]time.Time)

	go bcast.Transmitter(30000, tx)
	go bcast.Receiver(30000, rx)

	go SendHeartBeat(Elevator, tx)

	go ListenHeartBeat(rx, peerLastSeen)
}

func SendHeartBeat(elevator *config.Elevator, tx chan HeartBeat){
	for {
		tx <- HeartBeat{
			ElevatorID: elevator.ID,
			Timestamp: time.Now(),
		}
		time.Sleep(500*time.Millisecond)
	}
}

func ListenHeartBeat(rx chan HeartBeat, peerLastSeen map[string]time.Time){
	for {
		select {
		case hb := <- rx:
			peerLastSeen[hb.ElevatorID] = time.Now()
			fmt.Printf("Heartbeat recieved from %s\n", hb.ElevatorID)
		default: // Sjekker for offline heiser hvert 500ms
			time.Sleep(500 * time.Millisecond)
			for id, lastSeen := range peerLastSeen {
				if time.Since(lastSeen) > 3*time.Second { // Timeout på 1 sek
					fmt.Printf("Heis %s har gått offline!\n", id)
					delete(peerLastSeen, id) //
					}
				}
			}
		}
	}
