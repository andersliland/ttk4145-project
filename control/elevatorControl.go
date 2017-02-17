package control

import (
	"log"
	"os"
	"time"

	. "../driver"
	. "../utilities"
)

const watchdogTimeoutInterval = time.Second * 1
const watchdogKickInterval = watchdogTimeoutInterval / 3

func InitElevatorControl() {

	log.Println("From init")
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

func FSM(buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	localIP string) {

	wdog := time.NewTicker(watchdogTimeoutInterval)
	defer wdog.Stop()

	wdogKick := time.NewTicker(watchdogKickInterval)
	defer wdogKick.Stop()

	orderSlice := make([]ElevatorOrderMessage, 1)
	//	aliveElevators := make([]net.UDPAddr, 3)

	//var knownElevators = make(map[string])
	//var activeElevators = make(map[string]bool)

	for {
		select {
		case <-wdog.C:
			log.Println("[fsm] watchdog timeout. Kill process")
			os.Exit(1)

		case <-wdogKick.C:
			wdog = time.NewTicker(watchdogTimeoutInterval) // reset watchdog
			sendBackupChannel <- ElevatorBackupMessage{
				Time:     time.Now(),
				OriginIP: localIP,
				Event:    EvElevatorAliveMessage,
			}

		case b := <-buttonChannel: // Button handler, create order and broadcast to network
			//log.Println("[fsm] Received button from Floor:", b.Floor, ", Kind: ", b.Kind)
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

		case order := <-receiveOrderChannel:
			//log.Println("[fsm] Recieved event", order.Event)
			switch order.Event {
			case EvNewOrder:
				log.Println("[udp]", order)

				assignedOrder, err := ElevatorCostCalulation(order) // need to take in position of other elevators
				if err != nil {
					log.Println("[udp] ElevatorCostCalculation failed.")
				}
				orderSlice = append(orderSlice, order)
				//log.Println("Order Slice", orderSlice)
				//log.Println("[fsm] Assigned order to elevator: ", assignedOrder.AssignedTo)
				assignedOrder.Event = AckExecuteOrder
				sendMessageChannel <- assignedOrder // broadcast assigned order

			case AckExecuteOrder:
				if order.AssignedTo != localIP {
					order.Event = EvExecuteOrder
					sendMessageChannel <- order

				}

			case EvExecuteOrder:
				if order.AssignedTo == localIP {
					//TODO: make elevator execute order
					motorChannel <- MotorDown

				} //else idle/continue order

			case EvRestoreOrder:

			}

		}
	}
}

func ElevatorCostCalulation(newElevatorOrder ElevatorOrderMessage) (assignedOrder ElevatorOrderMessage, err error) {

	//TODO: calculate cost
	newElevatorOrder.AssignedTo = newElevatorOrder.OriginIP

	return newElevatorOrder, nil
}
