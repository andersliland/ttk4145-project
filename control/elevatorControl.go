package control

import (
	"log"
	"strconv"
	"time"

	"../cost"
	"../orders/"
	. "../utilities/"
)

const debugElevatorControl = true

func MessageLoop(
	newOrder chan bool,
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	broadcastOrderChannel chan OrderMessage,
	receiveOrderChannel chan OrderMessage,
	broadcastBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallHallOrderMatrix [NumFloors][2]ElevatorOrder,
	localIP string) {

	floorReached := make(chan int)
	go eventManager(newOrder, broadcastOrderChannel, floorReached, lightChannel, motorChannel, localIP)

	for {
		select {
		// Foreslår at kun buttonChannel og floorChannel er i denne filen
		// (det er kun det som er uavhengig av de andre heisene)

		case button := <-buttonChannel: // Hardware
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor+1))
			buttonHandler(button, broadcastOrderChannel, broadcastBackupChannel, lightChannel, motorChannel, OnlineElevators, ElevatorStatus, HallHallOrderMatrix, localIP)
			//newOrder <- true
		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			printElevatorControl("Elevator " + localIP + " reaced floor " + strconv.Itoa(floor+1))
			// Add cases for tickers
		}
	}
}

func buttonHandler(button ElevatorButton,
	broadcastOrderChannel chan<- OrderMessage,
	broadcastBackupChannel chan<- ElevatorBackupMessage,
	lightChannel chan<- ElevatorLight,
	motorChannel chan<- int,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallHallOrderMatrix [NumFloors][2]ElevatorOrder,
	localIP string) {

	switch button.Kind {
	case ButtonCallUp, ButtonCallDown:
		orderAssignedTo, err := cost.AssignOrderToElevator(button.Floor, button.Kind, OnlineElevators, ElevatorStatus, HallHallOrderMatrix)
		//printElevatorControl("Local assign order to " + orderAssignedTo)
		CheckError("[elevatorControl] Failed to assign Order to Elevator ", err)
		order := OrderMessage{
			Time:       time.Now(),
			Floor:      button.Floor,
			ButtonType: button.Kind,
			AssignedTo: orderAssignedTo,
			OriginIP:   localIP,
			SenderIP:   localIP,
			Event:      EventNewOrder,
		}
		broadcastOrderChannel <- order

	case ButtonCommand:
		orders.AddCabOrder(button, localIP)

		broadcastBackupChannel <- ElevatorBackupMessage{
				AskerIP: localIP,
				Event:   EventElevatorBackup,
				State: Elevator{
					LocalIP: localIP,
					//LastFloor: <-floorChannel ,
					//	Direction: ,
					//	IsMoving: ,
					//	DoorStatus: ,
					// CabOrders[button.Floor]: true, // why does this not work
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
		//motorChannel <- MotorStop
		//lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
		log.Println("Stop button pressed. Elevator will come to a halt.")
		//time.Sleep(1 * time.Second)
		//lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
		//os.Exit(1)
	}

}

func printElevatorControl(s string) {
	if debugElevatorControl {
		log.Println("[elevatorControl] \t", s)
	}
}
