package control

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"../orders/"
	. "../utilities/"
)

const debugElevatorControl = false

func MessageLoop(
	newOrder chan bool,
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	broadcastOrderChannel chan OrderMessage,
	receiveOrderChannel chan OrderMessage,
	broadcastBackupChannel chan BackupMessage,
	receiveBackupChannel chan BackupMessage,
	OnlineElevators map[string]bool,
	ElevatorStatus map[string]*Elevator,
	HallOrderMatrix [NumFloors][2]HallOrder,
	localIP string) {

	floorReached := make(chan int)
	go eventManager(newOrder, broadcastOrderChannel, broadcastBackupChannel, floorReached, lightChannel, motorChannel, localIP)
	time.Sleep(1 * time.Second)
	go setPanelLights(lightChannel, localIP)

	for {
		select {
		case button := <-buttonChannel: // Hardware
			printElevatorControl("New button push from " + localIP + " of type '" + ButtonType[button.Kind] + "' at floor " + strconv.Itoa(button.Floor+1))
			switch button.Kind {
			case ButtonCallUp, ButtonCallDown:
				if _, ok := OnlineElevators[localIP]; !ok {
					log.Println("[elevatorControl]\t Elevator offline, cannot accept new order")
				} else {
					orderAssignedTo, _ := orders.AssignOrderToElevator(button.Floor, button.Kind, OnlineElevators, ElevatorStatus)
					//log.Println("Order at floor " + strconv.Itoa(button.Floor+1) + " of type " + ButtonType[button.Kind] + " is assigned to " + orderAssignedTo)

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
				/*
					broadcastBackupChannel <- BackupMessage{
						AskerIP: localIP,
						Event:   EventElevatorBackup,
						Cab: CabOrder{
							Floor: button.Floor,
						},
					}
				*/
				orders.AddCabOrder(button, localIP)
				ElevatorStatus[localIP].SaveToFile("backupElevator")
				newOrder <- true
				//log.Println(ElevatorStatus[localIP])

			case ButtonStop:
				motorChannel <- Stop
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: true}
				log.Println("[elevatorControl]\t Stop button pressed. Elevator will come to a halt.")
				log.Println("[elevatorControl]\t You need to restart system to init elevator again.")
				time.Sleep(50 * time.Millisecond)
				lightChannel <- ElevatorLight{Kind: ButtonStop, Active: false}
				//os.Exit(1)

			}
		case floor := <-floorChannel: // Hardware
			floorReached <- floor
			fmt.Print(ColorYellow)
			log.Println("Elevator "+localIP+" reached floor "+strconv.Itoa(floor+1), ColorNeutral)
		}
	}
}

func setPanelLights(lightChannel chan ElevatorLight, localIP string) {
	var cabPanelLights [NumFloors]bool
	var hallPanelLights [NumFloors][2]bool
	for {
		for f := 0; f < NumFloors; f++ {
			if ElevatorStatus[localIP].CabOrders[f] == true && cabPanelLights[f] != true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: true}
				cabPanelLights[f] = true
				//printElevatorControl("Set CabOrder light on floor " + strconv.Itoa(f+1) + " on elevator " + localIP)
			} else if ElevatorStatus[localIP].CabOrders[f] == false && cabPanelLights[f] == true {
				lightChannel <- ElevatorLight{Floor: f, Kind: ButtonCommand, Active: false}
				cabPanelLights[f] = false
				//printElevatorControl("Clear CabOrder light on floor " + strconv.Itoa(f+1) + " on elevator " + localIP)
			}
			for k := ButtonCallUp; k <= ButtonCallDown; k++ {
				if (HallOrderMatrix[f][k].Status == Awaiting || HallOrderMatrix[f][k].Status == UnderExecution) && hallPanelLights[f][k] != true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: true}
					hallPanelLights[f][k] = true
					//printElevatorControl("Set HallOrder light on floor " + strconv.Itoa(f+1) + " of kind " + MotorStatus[] + " on elevator " + localIP)
				} else if (HallOrderMatrix[f][k].Status == NotActive) && hallPanelLights[f][k] == true {
					lightChannel <- ElevatorLight{Floor: f, Kind: k, Active: false}
					hallPanelLights[f][k] = false
					//printElevatorControl("Clear HallOrder light on floor " + strconv.Itoa(f+1) + " of kind " + MotorStatus[k+1] + " on elevator " + localIP)
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func printElevatorControl(s string) {
	if debugElevatorControl {
		log.Println("[elevatorControl]\t", s)
	}
}
