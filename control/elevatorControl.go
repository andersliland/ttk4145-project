package control

import (
	"log"
	"strconv"
	"time"

	. "../utilities/"
)

const debugElevatorControl = true

func MessageLoop(
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	WorkingElevators map[string]bool,
	RegisteredElevators map[string]*Elevator,
	HallOrderMatrix [NumFloors][2]ElevatorOrder,
	localIP string) {

	//newOrder := make(chan bool)
	floorReached := make(chan int)
	//go eventManager(newOrder, floorReached, lightChannel, motorChannel)

	for {
		select {
		//case message := <-receiveBackupChannel: // Network
		//case message := <-receiveOrderChannel: // Orders
		//	newOrder <- true
		//case message := <-timeOutChannel: // Timeout
		case button := <-buttonChannel: // Hardware
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor))
			buttonHandler(button, sendMessageChannel, sendBackupChannel, lightChannel, motorChannel, WorkingElevators, RegisteredElevators, HallOrderMatrix, localIP)

		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			// Add cases for tickers
		}
	}
}

func buttonHandler(button ElevatorButton,
	sendMessageChannel chan<- ElevatorOrderMessage,
	sendBackupChannel chan<- ElevatorBackupMessage,
	lightChannel chan<- ElevatorLight,
	motorChannel chan<- int,
	WorkingElevators map[string]bool,
	RegisteredElevators map[string]*Elevator,
	HallOrderMatrix [NumFloors][2]ElevatorOrder,
	localIP string) {

	//newOrder <- true
	switch button.Kind {
	case ButtonCallUp, ButtonCallDown:
		//orderAssignedTo, err := cost.AssignOrderToElevator(button.Floor, button.Kind, WorkingElevators, RegisteredElevators, HallOrderMatrix)
		//CheckError("[elevatorControl] Failed to assign Order to Elevator ", err)
		/*
			newOrder := ElevatorOrderMessage{
				Time:       time.Now(),
				Floor:      button.Floor,
				ButtonType: button.Kind,
				AssignedTo: orderAssignedTo,
				OriginIP:   localIP,
				SenderIP:   localIP,
				Event:      EventNewOrder,
			}
		*/
		//sendMessageChannel <- newOrder

	// Broadcast CabOrder as BackupMessage
	// Add LocalOrder to Execution
	case ButtonCommand:
		//queue.AddLocalOrder(button)
		// AddLocalOrder + SaveOrderToFile

		//RegisteredElevators[localIP].AddInternalOrder(button.Floor)

		sendBackupChannel <- ElevatorBackupMessage{
			AskerIP: localIP,
			Event:   EventElevatorBackup,
			State: ElevatorState{
				LocalIP: localIP,
				// LastFloor: ,
				//	Direction: ,
				//	IsMoving: ,
				//	DoorStatus: ,
				// CabOrders[button.Floor]: true, // why does this not work
				//CabButtonFloor:       button.Floor,
				//CabOrderMap[localIP]: button.Floor,
			},
			Cab: CabOrder{
				LocalIP:  localIP,
				OriginIP: localIP,
				Floor:    button.Floor,
				//ConfirmedBy: ,
				Timer: time.Now(),
			},
		}

	case ButtonStop:
		motorChannel <- MotorStop
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
		log.Println("Stop button pressed. Elevator will come to a halt.")
		time.Sleep(1 * time.Second)
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
		//os.Exit(1)
	}

}

func printElevatorControl(s string) {
	if debugSystemControl {
		log.Println("[elevatorControl]", s)
	}
}
