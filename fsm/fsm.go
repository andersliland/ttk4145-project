package fsm

import (
	. "../config"
	"../cost"
	. "../driver"
	//"../network"
	"log"
	"time"
	//"../watchdog"
)

const (
	idle = iota
	moving
	doorOpen
)

const watchdogTimeoutInterval = time.Second * 1
const watchdogKickInterval = watchdogTimeoutInterval / 3

func InitFSM() {

}

// ORDERS
func FSM(buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	costOrderChannel chan ElevatorOrderMessage,
	localIP string) {

	wdog := time.NewTicker(watchdogTimeoutInterval)
	defer wdog.Stop()

	wdogKick := time.NewTicker(watchdogKickInterval)
	defer wdogKick.Stop()

	for {
		select {
		case <-wdog.C:
			//log.Println("watchdog timeout")
			// implement timeout handling
			//os.Exit(1)

		case <-wdogKick.C:
			//log.Println("watchdog kick")
			// TODO: implement kick handling

			// Button handler, create order and broadcast to nettwork
		case b := <-buttonChannel:
			log.Println("[fsm] Recieved button from Floor:", b.Floor, ", Kind: ", b.Kind)
			switch b.Kind {

			case ButtonCallUp, ButtonCallDown, ButtonCommand:
				newOrder := ElevatorOrderMessage{
					Floor:      b.Floor,
					ButtonType: b.Kind,
					AssignedTo: "none",
					OriginIP:   localIP,
					SenderIP:   localIP,
					Event:      EvNewOrder,
				}
				sendMessageChannel <- newOrder

			case ButtonStop:
				//TODO: add support for stop button in driver
				motorChannel <- MotorDown
			}

			if b.Floor == 0 && b.Kind == 2 {
				motorChannel <- 0
				//log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}
			if b.Floor == 1 && b.Kind == 2 {
				motorChannel <- 1
				//log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}
			if b.Floor == 2 && b.Kind == 2 {
				motorChannel <- 2
				//log.Println("Button", "Floor:", b.Floor, "Kind:", b.Kind)
				lightChannel <- ElevatorLight{Floor: b.Floor, Kind: b.Kind, Active: true}
			}

		case f := <-floorChannel:
			if f != -1 {
				motorChannel <- 0
			}

		case order := <-receiveOrderChannel:
			switch order.Event {
			case EvNewOrder:
				assignedOrder, err := cost.ElevatorCostCalulation(order)
				if err != nil {
					log.Println("[udp] ElevatorCostCalculation failed.")
				}
				//log.Println("[fsm] Assigned order to elevator: ", assignedOrder.AssignedTo)
				assignedOrder.Event = EvExecuteOrder
				sendMessageChannel <- assignedOrder // broadcast assigned order

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
