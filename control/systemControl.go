package control

// This module is responsible for the communication between all the elevators, organising the
// order matrix and queue, and delegating orders to spesific elevators.
//
// Orders are organised in a local and remote list. Each list contains

// The system support two types of messages
// send

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"../orders"

	. "../utilities"
)

var debugSystemControl = true

func Init(localIP string) {
	ElevatorStatus[localIP] = ResolveElevator(Elevator{LocalIP: localIP})
}

func SystemControl(
	motorChannel chan<- int,
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
	var orderTimeout = OrderTimeout*time.Second + time.Duration(r.Intn(2000))*time.Millisecond // random timeout to prevent all elevator from timing out at the same time

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
					printSystemControl("Received EventElevatorOnline from a new elevator with IP " + backup.ResponderIP)
					ElevatorStatus[backup.ResponderIP] = ResolveElevator(backup.State)
				}
				updateOnlineElevators(ElevatorStatus, OnlineElevators, localIP, watchdogLimit)

			case EventElevatorBackup:
				//log.Println("Received an EventElevatorBackup from " + backup.AskerIP)
				//if backup.AskerIP != localIP { // shoud be !=
				//	ElevatorStatus[backup.AskerIP].UpdateElevatorStatus(backup)
				//}

			case EventRequestBackup:
				if backup.AskerIP == localIP { // TODO change to !=
					printSystemControl("Received an EventRequestBackup from " + backup.AskerIP)
					if _, ok := ElevatorStatus[backup.AskerIP]; ok {
						broadcastBackupChannel <- BackupMessage{
							AskerIP:         backup.AskerIP,
							ResponderIP:     localIP,
							Event:           EventBackupReturned,
							State:           *ElevatorStatus[backup.AskerIP],
							HallOrderMatrix: HallOrderMatrix,
						}
						printSystemControl("Broadcasting elevator state from elevator " + localIP)
					} else {
						log.Println("[systemControl]\t No stored state for elevator " + backup.AskerIP)
					}
				}

				// Restore state of elevator
			case EventBackupReturned:
				printSystemControl("Received EventBackupReturned from " + backup.ResponderIP)
				if backup.AskerIP == localIP {
					//ElevatorStatus[localIP] ResolveElevator()
					log.Printf("[systemControl]\t Received EventBackupReturned requested by me")

				} else {
					log.Printf("[systemControl]\t Received EventBackupReturned NOT requested by me")
				}
			default:
				log.Println("[systemControl]\tReceived invalid BackupMessage from", backup.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			//printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			switch order.Event {

			case EventNewOrder:
				if order.SenderIP == localIP {
					//printSystemControl("case: EventAckNewOrder")
				}
				fmt.Print(ColorGreen)
				log.Println("[systemControl]\t Order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1) + " is assigned to " + order.AssignedTo + ColorNeutral)

				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo //assume cost func is correct
				HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting             // set to NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()            // create new instance of ConfirmedBy map

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckNewOrder,
				}

				if order.OriginIP == localIP {
					printSystemControl("Starting ack timer on new order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("[systemControl]\t ACK-TIMEOUT\t newOrder at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
						timeoutChannel <- ExtendedHallOrder{
							Floor:        order.Floor,
							ButtonType:   order.ButtonType,
							TimeoutState: TimeoutAckNewOrder,
							Order: HallOrder{
								AssignedTo: order.AssignedTo,
							},
						}
					})
				}
			case EventAckNewOrder:
				//printSystemControl("case: EventAckNewOrder")
				// OriginIP is responsible for registering ack from other elevators
				if order.OriginIP == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
						printSystemControl("All elevators have ack'ed NewOrder at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy()

						broadcastOrderChannel <- OrderMessage{
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventOrderConfirmed,
						}
					}
				}

			case EventOrderConfirmed:
				if order.AssignedTo == localIP {
					newOrder <- true
				}

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderConfirmed,
				}

				// All other elevators than OriginIP start timer on order execution.
				// The elevator which the order is assigned to must timeout before the others.
				if order.OriginIP != localIP {
					printSystemControl("Start execution timeout on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					timeout := orderTimeout
					if order.AssignedTo != localIP {
						log.Println("THE MADMAN")
						timeout = 2 * orderTimeout
					}
					printSystemControl("Starting execution timer [EventOrderConfirmed] on order " + ButtonType[order.ButtonType] + " on floor " + strconv.Itoa(order.Floor+1))
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
						log.Println("Timeout\t orderUnderExecution - elevator " + order.OriginIP + " could not execute order (OriginIP != localIP)")
						timeoutChannel <- ExtendedHallOrder{
							Floor:        order.Floor,
							ButtonType:   order.ButtonType,
							OriginIP:     order.OriginIP,
							TimeoutState: TimeoutOrderExecution,
							Order: HallOrder{
								AssignedTo: order.AssignedTo,
							},
						}
					})
				}

			case EventAckOrderConfirmed:
				if order.SenderIP == localIP {
					//printSystemControl("case: EventAckOrderConfirmed")
				}
				if order.OriginIP == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
					if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
						printSystemControl("All elevators have ack'ed OrderConfirmed at Floor " + strconv.Itoa(order.Floor+1) + " of  type " + ButtonType[order.ButtonType])
						HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
						HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)
						timeout := orderTimeout
						if order.AssignedTo != localIP {
							timeout = 2 * orderTimeout
						}

						log.Println("[systemConrtol]\t OriginIP start execution timer [EventOrderConfirmed] on order "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)
						HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(timeout, func() {
							log.Println("Timeout\t\t orderUnderExecution - Elevator could not execute order (OriginIP == localIP)")
							timeoutChannel <- ExtendedHallOrder{
								Floor:        order.Floor,
								ButtonType:   order.ButtonType,
								OriginIP:     order.OriginIP,
								TimeoutState: TimeoutOrderExecution,
								Order: HallOrder{
									AssignedTo: order.AssignedTo,
								},
							}
						})
					}
				}

			case EventOrderCompleted:
				// This case is only sent from the eventManager after it detects that an order is completed.
				printSystemControl("case: EventOrderCompleted at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " for " + order.AssignedTo)
				HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = ""
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer() // stops timer set in EventAckOrderConfirmed
				//log.Println("[systemControl]\t EventOrderComplete stop timer at order  "+ButtonType[order.ButtonType]+" on floor "+strconv.Itoa(order.Floor+1)+" Timer: ", HallOrderMatrix[order.Floor][order.ButtonType].Timer)

				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: order.AssignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventAckOrderCompleted,
				}

				if order.AssignedTo == localIP {
					HallOrderMatrix[order.Floor][order.ButtonType].Timer = time.AfterFunc(ackTimeLimit, func() {
						log.Println("[systemControl]\t ACK-TIMEOUT\t orderCompleted at floor " + strconv.Itoa(order.Floor+1) + " for " + ButtonType[order.ButtonType] + " not ACK'ed by all ")
						broadcastOrderChannel <- OrderMessage{ // Should we send to timeoutChannel - or just resend OrderMessage?
							Floor:      order.Floor,
							ButtonType: order.ButtonType,
							AssignedTo: order.AssignedTo,
							OriginIP:   order.OriginIP,
							SenderIP:   localIP,
							Event:      EventAckOrderCompleted,
						}
					})
				}

			case EventAckOrderCompleted: // delete order from matrix and timer functions
				printSystemControl("case: EventAckOrderCompleted")
				HallOrderMatrix[order.Floor][order.ButtonType].ConfirmedBy[order.SenderIP] = true
				if allElevatorsHaveAcked(OnlineElevators, HallOrderMatrix, order) {
					fmt.Printf(ColorBlue)
					log.Println("[systemControl]\t All elevators have ack'ed OrderCompleted at Floor "+strconv.Itoa(order.Floor+1)+" of  type "+ButtonType[order.ButtonType], ColorNeutral)
					HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
					HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)

					resetTimerForAllAssignedOrders(order.Floor, orderTimeout, order.AssignedTo)
				}

			case EventReassignOrder:

				printSystemControl("case: EventReassignOrder")
				HallOrderMatrix[order.Floor][order.ButtonType].StopTimer()        // stop ackTimeout timer
				HallOrderMatrix[order.Floor][order.ButtonType].ClearConfirmedBy() // ConfirmedBy map an inner map (declared inside struct, and not initialized)
				HallOrderMatrix[order.Floor][order.ButtonType].Status = NotActive
				assignedTo, _ := orders.AssignOrderToElevator(order.Floor, order.ButtonType, OnlineElevators, ElevatorStatus)
				printSystemControl("case: EventReassignOrder assigned to" + assignedTo + ". OriginIP " + order.OriginIP)
				broadcastOrderChannel <- OrderMessage{
					Floor:      order.Floor,
					ButtonType: order.ButtonType,
					AssignedTo: assignedTo,
					OriginIP:   order.OriginIP,
					SenderIP:   localIP,
					Event:      EventNewOrder,
				}

			default:
				log.Println("Received an invalid OrderMessage from" + order.SenderIP)

			}
		case t := <-timeoutChannel:
			switch t.TimeoutState {
			case TimeoutAckNewOrder:
				log.Println("Not all elevators ACKed newOrder at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   localIP,
					Event:      EventNewOrder,
				}

			case TimeoutAckOrderConfirmed:
				log.Println("Not all elevators ACKed orderConfirmed at floor " + strconv.Itoa(t.Floor+1) + " for " + ButtonType[t.ButtonType] + ". Resending...")
				broadcastOrderChannel <- OrderMessage{
					Floor:      t.Floor,
					ButtonType: t.ButtonType,
					AssignedTo: t.Order.AssignedTo,
					SenderIP:   localIP,
					OriginIP:   t.OriginIP,
					Event:      EventOrderConfirmed,
				}

			case TimeoutOrderExecution: // EventAckOrderCompleted failed
				// kill self
				if t.Order.AssignedTo == localIP {
					motorChannel <- Stop
					time.Sleep(100 * time.Millisecond)
					fmt.Print(ColorRed)
					log.Fatal("SUICIDE, could not complete order "+ButtonType[t.ButtonType]+" at floor "+strconv.Itoa(t.Floor+1)+". OriginIP: "+t.OriginIP+" AssignedTo: "+t.Order.AssignedTo, ColorNeutral)
				}
				if t.OriginIP == localIP {
					printSystemControl("Order timeout I am OriginIP ")
					broadcastOrderChannel <- OrderMessage{
						Floor:      t.Floor,
						ButtonType: t.ButtonType,
						AssignedTo: t.Order.AssignedTo,
						OriginIP:   localIP, //randomly assign OriginIP for new order
						SenderIP:   localIP,
						Event:      EventReassignOrder,
					}

				} else {
					printSystemControl("Order timeout I am NOT OriginIP ")
					broadcastOrderChannel <- OrderMessage{
						Floor:      t.Floor,
						ButtonType: t.ButtonType,
						AssignedTo: t.Order.AssignedTo,
						OriginIP:   localIP, //randomly assign OriginIP for new order
						SenderIP:   localIP,
						Event:      EventReassignOrder,
					}

				}

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
				log.Println("[systemControl] \t All OnlineElevators", OnlineElevators, "Removed ", ElevatorStatus[k].LocalIP)

			}
		} else { // watchdog not timed out
			if OnlineElevators[k] != true {
				OnlineElevators[k] = true
				//printSystemControl("Added elevator " + ElevatorStatus[k].LocalIP + " in OnlineElevators")
				log.Println("[systemControl] \t All OnlineElevators", OnlineElevators, "Added ", ElevatorStatus[k].LocalIP)
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

func resetTimerForAllAssignedOrders(floor int, orderTimeout time.Duration, ip string) {
	// reset timer for all order AssignetTo == localIP
	for f := 0; f < NumFloors; f++ {
		for k := ButtonCallUp; k <= ButtonCallDown; k++ {
			if HallOrderMatrix[f][k].AssignedTo == ip {
				HallOrderMatrix[f][k].Timer.Reset(orderTimeout)
				log.Println("Reset timer on order" + ButtonType[k] + " at floor " + strconv.Itoa(f+1))
			}
		}
	}
}

func printSystemControl(s string) {
	if debugSystemControl {
		log.Println("[systemControl]\t", s)
	}
}
