# Elevator Project

Solution to the elevator project in TTK4145 from group 19. The project implements a distributed system for n elevators over m floors, with a strong emphasis on fault tolerance, efficiency and robustness. 

The project is written in Go (Golang) due to its support for concurrency via goroutines and channels. The elevator system uses a peer-to-peer architecture, where nodes communicate via UDP.  


## FAULT TOLERANCE:
- Motor power loss: motor timer
- Network: heartbeat messages
- Packet loss: redundancy 
- Software crash: backup 


## PROJECT OVERVIEW:
FSM: Handles the elevators behavior in response to events like button presses, arriving at floors, stop signals and updated orders from the network. It listens to input channels, updates the elevator state, and delegates logic decisions to the logic module. 

config: Contains shared constants, types and initialization functions that is used throughout the system. It holds elevator behaviour, request counters, and channels for communication between modules. The module ensures consistent data structures and configuration across all components. 

control: Handles the core logic of elevator movement and behaviour at each floor. It contains functions that decides when to stop, open the door and which direction to move next based on the current orders. The control module also manages the dor timer and obstruction logic. It updates all button and floor lights. 

elevio: Handles communication with the elevator hardware. It initializes the connection to the elevator server and provides functions for controlling motor direction, lights and door signals. It also polls hardware inputs such as button presses, floor sensors, obstruction switch and stop button. This modules acts as the hardware abstraction layer for the entire project. 

logic: Handles task assignment, request management and fault-tolerant coordination between elevators. It uses Hall Request Assigner to distribute hall calls based on the elevator behaviour. The module translates internal state to HRA format, update local orders and ensures synchronization across the network. A cyclic counter ensures consistent hall requests across peers.

Network: Manages the communication between elevators in a peer-to-peer setup. It handles broadcasting, tracks connected peers, detects peer loss and provides backup and recovery of orders. 

Timer: Provides non-blocking timers for controlling door open duration and detecting if the elevator motor is stuck (power loss) or if the door is obstructed. 


## HOW TO START THE PROJECT:
### Prerequisites
- Go must be installed and available in your system path.
- Required executables must be downloaded and placed in the correct location, as described below.

### Installation
- elevatorserver: https://github.com/TTK4145/elevator-server
- hall_request_assigner: Download from https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3, put the executable in the logic folder,    and run "chmod +x hall_request_assigner" in the terminal.

### Run
- "elevatorserver"
- "go run main.go -Id "your ID here"" (ID can be 1, 2, 3 and so on)

### Terminate 
- "ctrl + c" in terminal
