package control

import (
	"log"
	"time"

	. "../driver"
	. "../utilities"
)

func InitElevatorControl() {

}

func MessageLoop(
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	localIP string) {

	for {
		select {
		//case message := <-receiveBackupChannel: // Network
		//case message := <-receiveOrderChannel: // Orders
		//case message := <-timeOutChannel: // Timeout
		//case button := <-buttonChannel: // Hardware
		//case floor := <-floorChannel: // Hardware
		// Add cases for tickers
		}
	}
}

var externalOrderMatrix [NumFloors][NumButtons]ElevatorOrder

func FSM(buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	executeOrderChannel chan ElevatorOrderMessage,
	localIP string) {

	for {
		select {

		case b := <-buttonChannel: // Button handler, create order and broadcast to network
			//log.Println("[fsm] Received button from Floor:", b.Floor, ", ButtonType: ", b.ButtonType)
			switch b.Kind {
			case ButtonCallUp, ButtonCallDown, ButtonCommand:
				newOrder := ElevatorOrderMessage{
					Time:       time.Now(),
					Floor:      b.Floor,
					ButtonType: b.Kind,
					AssignedTo: "none",
					OriginIP:   localIP,
					SenderIP:   localIP,
					Event:      EvNewOrder,
				}
				sendMessageChannel <- newOrder

			case ButtonStop:
				motorChannel <- MotorStop
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("Stop button pressed. Elevator will come to a halt.")

			}

		case <-executeOrderChannel:
			//printDebug("Recieved an " + EventType[b.Event] + " from " + b.SenderIP + " with OriginIP " + b.OriginIP)
			// EvNewOrder
			//EvAckNewOrder
			//EvOrderConfirmed
			//EvAckOrderConfirmed
			//EvOrderDone
			//EvAckOrderDone
			//EvReasignOrder
			/*
				if b.Floor == Floor1 && b.ButtonType == ButtonCallUp {
					motorChannel <- MotorDown
					//log.Println("Button", "Floor:", b.Floor, "ButtonType:", b.ButtonType)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.ButtonType, Active: true}
				}

				if b.Floor == Floor2 && b.ButtonType == ButtonCallUp {
					motorChannel <- MotorStop
					//log.Println("Button", "Floor:", b.Floor, "ButtonType:", b.ButtonType)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.ButtonType, Active: true}
				}

				if b.Floor == Floor3 && b.ButtonType == ButtonCallUp {
					motorChannel <- MotorUp
					//log.Println("Button", "Floor:", b.Floor, "ButtonType:", b.ButtonType)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.ButtonType, Active: true}
				}
			*/

		}
	}
}

func ElevatorCostCalulation(newElevatorOrder ElevatorOrderMessage) (assignedOrder ElevatorOrderMessage, err error) {

	//TODO: calculate cost
	newElevatorOrder.AssignedTo = newElevatorOrder.OriginIP

	return newElevatorOrder, nil
}
