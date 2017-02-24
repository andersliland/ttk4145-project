package control

import (
	"log"
	"time"

	"github.com/andersliland/ttk4145-project/queue"

	. "../utilities/"
)

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
			log.Println("[elevatorControl] Received new button push")
			buttonHandler(button, sendMessageChannel, lightChannel, motorChannel, localIP)
			newOrder <- true
		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			// Add cases for tickers
		}
	}
}

func buttonHandler(button ElevatorButton, sendMessageChannel chan<- ElevatorOrderMessage,
	lightChannel chan<- ElevatorLight, motorChannel chan<- int, localIP string) {
	switch button.Kind {
	case ButtonCallUp, ButtonCallDown:
		newOrder := ElevatorOrderMessage{
			Time:       time.Now(),
			Floor:      button.Floor,
			ButtonType: button.Kind,
			AssignedTo: "none",
			OriginIP:   localIP,
			SenderIP:   localIP,
			Event:      EvNewOrder,
		}

		log.Println("[elevatorControl] Create new order", newOrder)
		sendMessageChannel <- newOrder

	case ButtonCommand:
		queue.AddLocalOrder(button)
		// AddLocalOrder + SaveOrderToFile
	case ButtonStop:
		motorChannel <- MotorStop
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
		log.Println("Stop button pressed. Elevator will come to a halt.")
		time.Sleep(time.Second)
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
		//os.Exit(1)
	}
}
