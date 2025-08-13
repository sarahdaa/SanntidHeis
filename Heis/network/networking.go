package network

import (
	"G19_heis2/Heis/config"
	"G19_heis2/Heis/elevio"
	"G19_heis2/Heis/logic"
	"G19_heis2/Heis/network/bcast"
	"G19_heis2/Heis/network/peers"
	"fmt"
	"reflect"
	"time"
)

func Networking(
	elevator 						*config.Elevator,
	networkChannels 				*config.NetworkChannels,
	lightChannels 					*config.LightChannels,
) {
	var numberOfPeers int
	peerTimeout := 3 * time.Second

	peersPort := 20034
	bcastPort := 30073

	networkChannels.StateRX = make(chan config.Elevator)
	networkChannels.StateTX = make(chan config.Elevator)

	networkChannels.PeerUpdate = make(chan peers.PeerUpdate, config.Buffsize)

	peerTxEnable := make(chan bool, config.Buffsize)
	peerTxEnable <- true
	peerLastSeen := make(map[string]time.Time, config.Buffsize)

	backupTx := make(chan map[string]config.Elevator, config.Buffsize)
	backupRx := make(chan map[string]config.Elevator, config.Buffsize)
	backupAckTx := make(chan string, config.Buffsize)
	backupAckRx := make(chan string, config.Buffsize)

	elevatorMap := make(map[string]config.Elevator)
	hallRequestlist := make(map[string][][]config.Counter)
	backupMap := make(map[string]config.Elevator)

	go peers.Transmitter(peersPort, elevator.ID, peerTxEnable)
	go peers.Receiver(peersPort, networkChannels.PeerUpdate)

	go bcast.Transmitter(bcastPort, networkChannels.StateTX, backupTx, backupAckTx)
	go bcast.Receiver(bcastPort, networkChannels.StateRX, backupRx, backupAckRx)

	go SendStateUpdate(elevator, networkChannels.StateTX)

	ticker := time.NewTicker(100 * time.Millisecond)

	hallRequestlist[elevator.ID] = elevator.HallRequests
	syncedElevators := false

	for {
		select {

		case backupData := <-backupRx:
			for id, backupElevator := range backupData {
				if id == elevator.ID {
					if (len(backupElevator.MyOrders) == config.NumFloors) && 
						elevator.HallRequests[0][elevio.BT_HallUp] == config.Initial{

						for floor := 0; floor < config.NumFloors; floor++ {
							elevator.MyOrders[floor][elevio.BT_Cab] = backupElevator.MyOrders[floor][elevio.BT_Cab]
						}
					}
					for i := 0; i < 10; i ++ {
						backupAckTx <- elevator.ID
					}
					
					networkChannels.UpdatedMyOrders <- elevator.MyOrders
				}
			}


		case elevatorID := <-backupAckRx:
			delete(backupMap, elevatorID)


		case receivedState := <-networkChannels.StateRX:
			
			peerLastSeen[receivedState.ID] = time.Now()
			numberOfPeers = len(peerLastSeen)

			elevatorMap[receivedState.ID] = receivedState

			hallRequestlist[elevator.ID] = elevator.HallRequests

			if elevator.ID != receivedState.ID {
				hallRequestlist[receivedState.ID] = receivedState.HallRequests
			}

			updatedHallRequestlist := logic.CyclicCounter(hallRequestlist, elevator, numberOfPeers)

			if allPeersHaveSameRequestList(updatedHallRequestlist) {
				if !syncedElevators {
					syncedElevators = true
				}
			} else {
				if syncedElevators {
					syncedElevators = false
				}
			}

			if syncedElevators {
				HRAInputRequests := logic.CreateHRAInputRequests(elevator.HallRequests)
				for _, peer := range elevatorMap{
					if !peer.MotorPower {
						delete(elevatorMap, peer.ID)
					}
				}
				HRAInputStates := logic.ElevatorToHRAELEV(elevatorMap)
				logic.UpdateElevatorHallRequests(elevator, HRAInputRequests, HRAInputStates, networkChannels, lightChannels)
				}
				

		case peerUpdate := <-networkChannels.PeerUpdate:

			fmt.Println("Peer update:", peerUpdate.Peers)
			fmt.Println("New: ", peerUpdate.New)
			fmt.Println("Lost: ", peerUpdate.Lost)

			if _, exists := backupMap[peerUpdate.New]; exists {
				for i := 0; i < 100; i++ {
					backupTx <- backupMap
				}
			}

			for _, lostPeer := range peerUpdate.Lost {
				fmt.Printf("Heis %s har gått offline!\n", lostPeer)

				backupMap[lostPeer] = elevatorMap[lostPeer]
				delete(peerLastSeen, lostPeer)
				delete(elevatorMap, lostPeer)
				delete(hallRequestlist, lostPeer)
			}

			fmt.Println("peerUpdate.lost: ", peerUpdate.Lost)

		case <-ticker.C:
			for id, lastSeen := range peerLastSeen {
				if time.Since(lastSeen) > peerTimeout {
					fmt.Printf("Heis %s har vært borte i over %.0f sekunder! \n", id, peerTimeout.Seconds())
					delete(peerLastSeen, id)

				}
			}
		}
	}
}

func SendStateUpdate(
	elevator 			*config.Elevator,
	txElevator 			chan config.Elevator,
	) {

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elevCopy := logic.DeepCopyElevator(*elevator)
			for i := 0; i < 3; i++ {
				txElevator <- elevCopy
			}

		}
	}
}


func allPeersHaveSameRequestList(requestsMap map[string][][]config.Counter) bool {
	if len(requestsMap) == 0 {
		return false
	}

	var firstPeerRequests [][]config.Counter
	for _, req := range requestsMap {
		firstPeerRequests = req
		break
	}

	for _, req := range requestsMap {
		if !reflect.DeepEqual(firstPeerRequests, req) {
			return false
		}
	}
	return true
}
