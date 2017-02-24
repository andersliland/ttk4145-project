package control

import (
	"log"
	"time"

	. "../utilities/"
)

const debugElevatorControl = false

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
			//log.Println("[elevatorControl] Received new button push")
			buttonHandler(button, sendMessageChannel, sendBackupChannel, lightChannel, motorChannel, localIP)
			//newOrder <- true

		case floor := <-floorChannel: // Hardware
			//floorHandler(floor)
			floorReached <- floor
			// Add cases for tickers
		}
	}
}

func buttonHandler(button ElevatorButton, sendMessageChannel chan<- ElevatorOrderMessage, sendBackupChannel chan<- ElevatorBackupMessage,
	lightChannel chan<- ElevatorLight, motorChannel chan<- int, localIP string) {

	//var cabOrderMaps = make(map[string]*Elevator) // containing last known state
	//cabOrderMaps[localIP] = ResolveElevatorState(ElevatorState{LocalIP: localIP})

	switch button.Kind {
	case ButtonCallUp, ButtonCallDown:
		log.Println("[elevatorControl] Received HallButton push")

		newOrder := ElevatorOrderMessage{
			Time:       time.Now(),
			Floor:      button.Floor,
			ButtonType: button.Kind,
			AssignedTo: "none",
			OriginIP:   localIP,
			SenderIP:   localIP,
			Event:      EvNewOrder,
		}
		sendMessageChannel <- newOrder

	case ButtonCommand:
		log.Println("[elevatorControl] Received CabButton push")

		// if elevator is not moving && last floor == current floor
		// else
		// add cabOrder to local map

		//queue.AddLocalOrder(button)
		// AddLocalOrder + SaveOrderToFile
		//	log.Printf("cabOrder type %T ", cabOrdersMap[localIP])

		sendBackupChannel <- ElevatorBackupMessage{
			AskerIP:     localIP,
			ResponderIP: string(button.Floor),
			Event:       EvBackupState,
			State: ElevatorState{
				LocalIP: localIP,
				//LastFloor: ,
				//	Direction: ,
				//	IsMoving: ,
				//	DoorStatus: ,
				//CabOrders[button.Floor]: true, // why does this not work
				CabOrderInt: button.Floor,
			},
		}

		log.Println("[elevatorControl] Send CabButton sync message")

	case ButtonStop:
		motorChannel <- MotorStop
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
		log.Println("Stop button pressed. Elevator will come to a halt.")
		time.Sleep(1 * time.Second)
		lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
		//os.Exit(1)
	}
}
