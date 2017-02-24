package control

import (
	"log"
	"os"
	"time"

	"../queue"
	. "../utilities/"
)

const watchdogTimeoutInterval = time.Second * 1
const watchdogKickInterval = watchdogTimeoutInterval / 3

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

	newOrder := make(chan bool)
	floorReached := make(chan int)
	go eventManager(newOrder, floorReached, lightChannel, motorChannel)
	log.Println("SUCCESS [elevatorControl]: Initialization") // remove

	for {
		select {
		//case message := <-receiveBackupChannel: // Network
		//case message := <-receiveOrderChannel: // Orders
		//	newOrder <- true
		//case message := <-timeOutChannel: // Timeout
		case button := <-buttonChannel: // Hardware
			buttonHandler(button, sendMessageChannel, lightChannel, motorChannel)
			newOrder <- true
		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			// Add cases for tickers
		}
	}
}

func buttonHandler(button ElevatorButton, sendMessageChannel chan ElevatorOrderMessage,
	lightChannel chan ElevatorLight, motorChannel chan int) {
	switch button.Kind {
	case ButtonCallUp, ButtonCallDown:
		newOrder := ElevatorOrderMessage{
			Time:       time.Now(),
			Floor:      button.Floor,
			ButtonType: button.Kind,
			AssignedTo: "none",
			//OriginIP:   localIP,
			//SenderIP:   localIP,
			//Event: EvNewOrder,
		}
		sendMessageChannel <- newOrder
	case ButtonCommand:
		queue.AddLocalOrder(button)
		// AddLocalOrder + SaveOrderToFile
	case ButtonStop:
		motorChannel <- MotorStop
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
		log.Println("Stop button pressed. Elevator will come to a halt.")
		time.Sleep(time.Second)
		os.Exit(1)
	}
}

// --- //
// --- //
// --- //
func FSM(buttonChannel chan ElevatorButton,
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
		case b := <-buttonChannel: // Button handler, create order and broadcast to network
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

				if b.Floor == Floor1 && b.Kind == ButtonCallUp {
					motorChannel <- MotorDown
					log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
				}

				if b.Floor == Floor2 && b.Kind == ButtonCallUp {
					motorChannel <- MotorStop
					log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
				}

				if b.Floor == Floor3 && b.Kind == ButtonCallUp {
					motorChannel <- MotorUp
					log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
					lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
				}

			case ButtonStop:
				motorChannel <- MotorStop
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("Stop button pressed. Elevator will come to a halt.")
				time.Sleep(time.Second)
				os.Exit(1)
			}

		case f := <-floorChannel:
			if f != -1 {
				motorChannel <- MotorStop
			}

		}

	}

}
