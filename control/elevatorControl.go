package control

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"../orders/"
	. "../utilities/"
)

const debugElevatorControl = false

func MessageLoop(
	newOrder chan bool,
	floorReached chan int,
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	broadcastOrderChannel chan OrderMessage,
	receiveOrderChannel chan OrderMessage,
	broadcastBackupChannel chan BackupMessage,
	receiveBackupChannel chan BackupMessage,
	orderCompleteChannel chan OrderMessage,
	onlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallOrderMatrix [NumFloors][2]HallOrder,
	localIP string) {

	var orderTimeout = OrderTimeout * time.Second

	go eventManager(newOrder, broadcastOrderChannel, broadcastBackupChannel, orderCompleteChannel, floorReached, lightChannel, motorChannel, localIP)
	time.Sleep(1 * time.Second)

	for {
		select {
		case button := <-buttonChannel: // Hardware
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor+1))
			switch button.Kind {
			case ButtonCallUp, ButtonCallDown:
				if _, ok := onlineElevators[localIP]; !ok {
					log.Println("[elevatorControl]\t Elevator offline, cannot accept new order")
				} else {
					orderAssignedTo, _ := orders.AssignOrderToElevator(button.Floor, button.Kind, onlineElevators, ElevatorStatus)
					broadcastOrderChannel <- OrderMessage{
						Floor:      button.Floor,
						ButtonType: button.Kind,
						AssignedTo: orderAssignedTo,
						OriginIP:   localIP,
						SenderIP:   localIP,
						Event:      EventNewOrder,
					}
				}
			case ButtonCommand:
				orders.AddCabOrder(button, localIP)
				if err := SaveBackup("backupElevator", ElevatorStatus[localIP].CabOrders); err != nil {
					log.Println("[elevatorControl]\t Save Backup failed: ", err)
				}
				resetTimerForAllAssignedOrders(orderTimeout, localIP) // reset timer on Cabbutton spamming
				fmt.Printf(ColorGreen)
				log.Println("[elevatorControl]\t CabOrder "+ButtonType[button.Kind]+"\ton floor "+strconv.Itoa(button.Floor+1), ColorNeutral)
				newOrder <- true

			case ButtonStop:
				motorChannel <- Stop
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("[elevatorControl]\t Stop button pressed. Elevator will come to a halt.")
				log.Println("[elevatorControl]\t You need to restart system to init elevator again.")
				time.Sleep(200 * time.Millisecond)
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
				os.Exit(1)

			}
		case floor := <-floorChannel: // Hardware
			floorReached <- floor
			fmt.Print(ColorYellow)
			log.Println("[elevatorControl]\t Elevator "+localIP+" reached floor "+strconv.Itoa(floor+1), ColorNeutral)
			resetTimerForAllAssignedOrders(orderTimeout, localIP)
		}
	}
}

func resetTimerForAllAssignedOrders(orderTimeout time.Duration, ip string) {
	// reset timer for all order AssignetTo == localIP
	for f := 0; f < NumFloors; f++ {
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[f][k].AssignedTo == ip {
				HallOrderMatrix[f][k].Timer.Reset(orderTimeout)
				//log.Println("[systemControl]\t Reset timer on order " + ButtonType[k] + " at floor " + strconv.Itoa(f+1))
			}
		}
	}
}

func printElevatorControl(s string) {
	if debugElevatorControl {
		log.Println("[elevatorControl]\t", s)
	}
}
