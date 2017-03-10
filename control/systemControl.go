package control

// This module is responsible for the communication between all the elevators, organising the
// order matrix and queue, and delegating orders to spesific elevators.
//
// Orders are organised in a local and remote list. Each list contains

// The system support two types of messages
// send

import (
	"log"
	"strconv"
	"time"

	. "../utilities"
)

var debugSystemControl = false

func Init(localIP string) {
	ElevatorStatus[localIP] = ResolveElevator(Elevator{LocalIP: localIP})
}

func SystemControl(
	newOrder chan bool,
	broadcastOrderChannel chan<- OrderMessage,
	receiveOrderChannel <-chan OrderMessage,
	broadcastBackupChannel chan<- BackupMessage,
	receiveBackupChannel <-chan BackupMessage,
	executeOrderChannel chan<- OrderMessage,
	localIP string) {

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond
	const ackTimeLimit = 500 * time.Millisecond

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)

	for {
		select {
		// Watchdog
		case <-watchdogKickTimer.C:
			broadcastBackupChannel <- ResolveWatchdogKickMessage(ElevatorStatus[localIP])
			//log.Printf("[systemControl] Watchdog send IAmAlive from %v \n", localIP)

		case <-watchdogTimer.C:
			updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)
			//log.Println("[systemControl] Active Elevators", OnlineElevators)

			// Network
		case backup := <-receiveBackupChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[backup.Event], backup.AskerIP)
			switch backup.Event {
			// resolved incomming alive, if timeout remove elevator
			case EventElevatorOnline:
				if _, ok := ElevatorStatus[backup.ResponderIP]; ok { // check if a value exsist for ResponderIP
					ElevatorStatus[backup.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					printSystemControl("Received EventElevatorOnline from a new elevator with IP" + backup.ResponderIP)
					ElevatorStatus[backup.ResponderIP] =
						ResolveElevator(backup.State)
				}
				updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)

			case EventElevatorBackup:
				printSystemControl("Received an EventElevatorBackup from " + backup.AskerIP)

			case EventRequestBackup:
				printSystemControl("Received an EventRequestBackup from " + backup.AskerIP)
				if backup.AskerIP != localIP {
					broadcastBackupChannel <- BackupMessage{
						AskerIP:     backup.AskerIP,
						ResponderIP: localIP,
						Event:       EventElevatorBackupReturned,
						//State:       Elevator{},
					}

				} else {
					printSystemControl("No stored state for elevator at selv " + localIP)
					/*
						broadcastBackupChannel <- BackupMessage{
							AskerIP:     backup.AskerIP,
							ResponderIP: localIP,
							Event:       EventElevatorBackupReturned,
							State:       ElevatorState{},
						}
					*/

				}

			case EventElevatorBackupReturned:
				printSystemControl("Received EventElevatorBackupReturned from " + backup.ResponderIP)
				if backup.AskerIP == localIP {
					// i requested this backup, update
					// HallOrderMatrix
					// CabHallOrderMatrix

				} else {
					log.Printf("[systemControl] Received EventElevatorBackupReturned not requested by me")

				}
			case EventCabOrder:
				printSystemControl("Received EventCabOrder from " + backup.AskerIP)
				if backup.AskerIP == localIP {
					printSystemControl("Received an EventCabOrder from selv, ignoring")
				} else {
					//CabHallOrderMatrix[backup.State.LastFloor].Status = Awaiting

				}

				//save order in map map[string] bool

			case EventAckCabOrder:

			default:
				log.Println("Received invalid BackupMessage from", backup.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			switch order.Event {

			case EventNewOrder:
				printSystemControl("Order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1) + ", assigned to " + order.AssignedTo)
				switch HallOrderMatrix[order.Floor][order.ButtonType].Status {
				case NotActive:
					printSystemControl("Received order in EventNewOrder, case NotActive")
					HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo //assume cost func is correct
					HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting
					HallOrderMatrix[order.Floor][order.ButtonType].InitConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

					// OriginIP is resposnible for order until it is assigned
					if order.OriginIP == localIP {
						// timeout handeling error where order is not acked
						HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
							log.Println("Timeout \t A new order was not acked by all ")
							/*
									ackTimeout <- HallOrder{
									 Floor: order.Floor,
									 Type: order.ButtonType,
									 Order: HallOrder{
										 Status: Awaiting,
										 AssignedTo: order.AssignedTo,
										 Timer: HallOrderMatrix[order.Floor][order.ButtonType].Timer,
									 }
								}
							*/
						})

					}

					broadcastOrderChannel <- OrderMessage{
						Floor:      order.Floor,
						ButtonType: order.ButtonType,
						AssignedTo: order.AssignedTo,
						OriginIP:   order.OriginIP,
						SenderIP:   localIP,
						Event:      EventAckNewOrder,
					}
				case Awaiting:
					printSystemControl("Received New order allready awaiting")
				case UnderExecution:
					printSystemControl("Received NewOrder allready executing")
				}
			case EventAckNewOrder:
				// OriginIP is responsible for register ack from other elevators
				if order.OriginIP == localIP {
					switch HallOrderMatrix[order.Floor][order.ButtonType].Status {
					case Awaiting:
						HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
						if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
							printSystemControl("All elevators have ack'ed order at Floor " + strconv.Itoa(order.Floor) + " of  type " + ButtonType[order.ButtonType])
							HallOrderMatrix[order.Floor][order.ButtonType].Timer.Stop() // stop ackTimeout timer
							//HallOrderMatrix[order.Floor][order.ButtonType]

							newOrder <- true

						} else {
							printSystemControl("Not all elevators acked")

						}

						broadcastOrderChannel <- OrderMessage{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventOrderConfirmed,
						}

					case UnderExecution:
						printSystemControl("Received EventAckNewOrder which is UnderExecution")

					case NotActive:
						printSystemControl("Received EventAckNewOrder which is NotActive")
					}
				}

			case EventOrderConfirmed:

				// the order is confirmed, start executing

			case EventAckOrderConfirmed:

			case EventOrderCompleted:
				//printSystemControl("Received EventAckNewOrder which is NotActive")

			case EventAckOrderCompleted:
				// delete order from matrix and timer functions

			default:
				printSystemControl("Received an invalid OrderMessage from" + order.SenderIP)

			}

		} // select
	} // for
} //function

// removes elevator from 'OnlineElevators' if watchdog timeout
// adds elevator to 'OnlineElevators' if watchdog not timeout
func updateOnlineElevators(ElevatorStatus map[string]*Elevator, OnlineElevators map[string]bool, localIP string, watchdogLimit time.Duration) {
	for k := range ElevatorStatus {
		if time.Since(ElevatorStatus[k].Time) > watchdogLimit { //watchdog timeout
			if OnlineElevators[k] == true {
				delete(OnlineElevators, k)
				printSystemControl("Removed elevator " + ElevatorStatus[k].LocalIP + " in OnlineElevators")
				log.Printf("[systemControl] \t All Working elevators %v", OnlineElevators)

			}
		} else { // watchdog not timed out
			if OnlineElevators[k] != true {
				OnlineElevators[k] = true
				printSystemControl("Added elevator " + ElevatorStatus[k].LocalIP + " in OnlineElevators")
				log.Printf("[systemControl] \t All OnlineElevators %v", OnlineElevators)
			}
		}
	}
}

func allElevatorsHaveAcked(OnlineElevators map[string]bool, HallOrderMatrix [NumFloors][2]HallOrder, order OrderMessage) bool {
	for ip, _ := range OnlineElevators {
		if _, confirmedBy := HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[ip]; !confirmedBy {
			return false
		}
	}
	return true

}
func printSystemControl(s string) {
	if debugSystemControl {
		log.Println("[systemControl]\t", s)
	}
}
