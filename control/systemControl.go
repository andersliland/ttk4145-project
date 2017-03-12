package control

// This module is responsible for the communication between all the elevators, organising the
// order matrix and queue, and delegating orders to spesific elevators.
//
// Orders are organised in a local and remote list. Each list contains

// The system support two types of messages
// send

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	. "../utilities"
)

var debugSystemControl = true

func Init(localIP string) {
	ElevatorStatus[localIP] = ResolveElevator(Elevator{LocalIP: localIP})
}

func SystemControl(
	newOrder chan<- bool,
	timeoutChannel chan ExtendedHallOrder,
	broadcastOrderChannel chan<- OrderMessage,
	receiveOrderChannel <-chan OrderMessage,
	broadcastBackupChannel chan<- BackupMessage,
	receiveBackupChannel <-chan BackupMessage,
	executeOrderChannel chan<- OrderMessage,
	localIP string) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond
	const ackTimeLimit = 500 * time.Millisecond
	var orderTimeout = 5*time.Second + time.Duration(r.Intn(2000))*time.Millisecond // random timeout to prevent all elevator from timing out at the same time

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)

	broadcastBackupChannel <- BackupMessage{
		AskerIP: localIP,
		Event:   EventRequestBackup,
	}

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
					ElevatorStatus[backup.ResponderIP] = ResolveElevator(backup.State)
				}
				updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)

			case EventElevatorBackup:
				//log.Println("Received an EventElevatorBackup from " + backup.AskerIP)

				if backup.AskerIP != localIP { // shoud be !=
					ElevatorStatus[backup.AskerIP].UpdateElevatorStatus(backup)
					//for k := range ElevatorStatus {
					//log.Println("LOCAL IP " + k + " - Floor " + strconv.Itoa(ElevatorStatus[k].Floor+1))
					//}

					//update ElevatorStatus
				}

			case EventRequestBackup:
				if backup.AskerIP == localIP { // TODO change to !=
					//printSystemControl("Received an EventRequestBackup from " + backup.AskerIP)
					if _, ok := ElevatorStatus[backup.AskerIP]; ok {
						broadcastBackupChannel <- BackupMessage{
							AskerIP:         backup.AskerIP,
							ResponderIP:     localIP,
							Event:           EventBackupReturned,
							State:           *ElevatorStatus[backup.AskerIP],
							HallOrderMatrix: HallOrderMatrix,
						}

						//printSystemControl("Broadcasting elevator state from elevator " + localIP)

					} else {
						printSystemControl("No stored state for elevator " + backup.AskerIP)
					}
				}

				// Restore state of elevator
			case EventBackupReturned:
				//printSystemControl("Received EventBackupReturned from " + backup.ResponderIP)
				if backup.AskerIP == localIP {
					for floor, hallOrdersAtFloor := range backup.HallOrderMatrix {
						for buttonKind, hallOrder := range hallOrdersAtFloor {
							if hallOrder.Status == UnderExecution && hallOrder.AssignedTo != localIP && HallOrderMatrix[floor][buttonKind].Status == NotActive {
								HallOrderMatrix[floor][buttonKind].Status = UnderExecution
								HallOrderMatrix[floor][buttonKind].ClearConfirmedBy()
								HallOrderMatrix[floor][buttonKind].AssignedTo = hallOrder.AssignedTo
								HallOrderMatrix[floor][buttonKind].Timer = time.AfterFunc(2*orderTimeout, func() {
									log.Println("[systemControl]\tAn order under execution timed out for elevator " + localIP)

								})

							}
						}
					}

				} else {
					log.Printf("[systemControl]\t Received EventBackupReturned not requested by me")
				}
			case EventCabOrder:
			case EventAckCabOrder:
			default:
				log.Println("[systemControl]\tReceived invalid BackupMessage from", backup.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			//printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			switch order.Event {

			case EventNewOrder:
				printSystemControl("Order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1) + ", assigned to " + order.AssignedTo)

				printSystemControl("Received order in EventNewOrder, case NotActive")
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo //assume cost func is correct
				HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

				// OriginIP is resposnible for order until it is assigned
				if order.OriginIP == localIP {
					//printSystemControl("Starting timeoutTimer [EventNewOrder] on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					// timeout handeling error where order is not acked
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("Timeout \t A new order was not ACKed by all ")
						timeoutChannel <- ExtendedHallOrder{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							Order: HallOrder{
								Status:     NotActive,
								AssignedTo: order.AssignedTo,
								Timer:      HallOrderMatrix[order.Floor][order.ButtonType].Timer,
							},
						}
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

			case EventAckNewOrder:
				// OriginIP is responsible for register ack from other elevators
				if order.OriginIP == localIP {
					switch HallOrderMatrix[order.Floor][order.ButtonType].Status {
					case Awaiting:
						HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
						if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
							printSystemControl("All elevators have ack'ed order at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
							HallOrderMatrix[order.Floor][order.ButtonType].Timer.Stop() // stop ackTimeout timer
							newOrder <- true
						} else {
							log.Println("[systemControl]\tNot all elevators ACK newOrder")
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
			case EventAckOrderConfirmed:
				// ack AssignedTo from all OnlineElevators
				if order.AssignedTo == localIP {
					newOrder <- true
				}

			case EventOrderCompleted:
				//log.Println("[systemControl]\t Received EventOrderCompleted")
				printSystemControl("Floor: " + strconv.Itoa(order.Floor+1) + " ButtonType: " + ButtonType[order.ButtonType])

				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

				if order.OriginIP == localIP {
					//printSystemControl("Starting timeoutTimer [EventOrderCompleted] on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("[systemControl]\t Timeout \t A new OrderCompleted was not ACKed by all ")
						timeoutChannel <- ExtendedHallOrder{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							Order: HallOrder{
								Status:     NotActive,
								AssignedTo: order.AssignedTo,
								Timer:      HallOrderMatrix[order.Floor][order.ButtonType].Timer,
							},
						}
					})
				}

				//if order.SenderIP != localIP {
				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderCompleted,
				}
				//}

			case EventAckOrderCompleted: // delete order from matrix and timer functions
				//printSystemControl("Received EventAckOrderCompleted")
				if order.AssignedTo == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
						printSystemControl("All elevators have ACKed orderComplete at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrix[order.Floor][order.ButtonType].Timer.Stop() // stop ackTimeout timer
					} else {
						log.Println("[systemControl]\tNot all elevators ACK orderComplete")
					}
				}
			default:
				log.Println("Received an invalid OrderMessage from" + order.SenderIP)
			}

		case t := <-timeoutChannel:
			switch t.Order.Status {
			case NotActive: // EventAckNewOrder failed
				log.Println("Not all elevators ACKed newOrder. Resending")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   localIP,
					Event:      EventNewOrder,
				}

			case Awaiting: // EventAckrderConfirmed failed
				log.Println("Not all elevators ACKed OrderConfirmed. Resending")

			case UnderExecution: //

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
				//printSystemControl("Removed elevator " + ElevatorStatus[k].LocalIP + " in OnlineElevators")
				log.Printf("[systemControl] \t All OnlineElevators  %v", OnlineElevators)

			}
		} else { // watchdog not timed out
			if OnlineElevators[k] != true {
				OnlineElevators[k] = true
				//printSystemControl("Added elevator " + ElevatorStatus[k].LocalIP + " in OnlineElevators")
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
