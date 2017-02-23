package network

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	. "../utilities"
)

const debug = false

func Init(sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage) (localIP string, err error) {

	udpSendDatagramChannel := make(chan UDPMessage, 10)
	udpReceiveDatagramChannel := make(chan UDPMessage, 5)

	go sendMessageHandler(sendMessageChannel, sendBackupChannel, udpSendDatagramChannel)
	go receiveMessageHandler(receiveOrderChannel, receiveBackupChannel, udpReceiveDatagramChannel)

	localIP, err = InitUDP(udpSendDatagramChannel, udpReceiveDatagramChannel)
	CheckError("", err)

	return localIP, nil
}

// Receive message from main.go, marshal and send down to udp.go
func sendMessageHandler(sendMessageChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	udpSendDatagramChannel chan UDPMessage) {

	for {
		select {
		case message := <-sendMessageChannel:
			data, err := json.Marshal(message)
			//log.Println("[network] Message to be sendt", message)
			if err != nil {
				log.Println("ERROR [network]: sendMessage marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Type: 1, Data: data} // UDPMessage
				//log.Println("[network] Sent a Message with event", EventType[message.Event])

			}
		case message := <-sendBackupChannel:
			networkPack, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendBackup marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Type: 2, Data: networkPack} // UDPMessage
				//log.Printf("[network] Sent a BackupMessage from %v with event %v", message.AskerIP, EventType[message.Event])
			}
		}
	}
}

// Receive message from udp.go, unmarshal and send up to main
func receiveMessageHandler(
	receiveOrderChannel chan ElevatorOrderMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	udpReceiveDatagramChannel chan UDPMessage) {

	for {
		select {
		case msg := <-udpReceiveDatagramChannel:
			var f interface{}
			err := json.Unmarshal(msg.Data[:msg.Length], &f) // stores Unmarshal as maps in
			if err != nil {
				log.Println("ERROR [network]: Unmarshal failed", err)
			} else {
				m := f.(map[string]interface{})
				event := int(m["Event"].(float64)) // type assertion, float64 because			//log.Printf("event %T, %v", event, event)
				if event <= 3 && event >= 0 {
					var backupMessage = ElevatorBackupMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &backupMessage); err == nil { //unmarshal into correct message struct
						//log.Printf("[network] Received a backupMessage from %v with event %v \n", backupMessage.ResponderIP, EventType[backupMessage.Event])
						receiveBackupChannel <- backupMessage
					} else {
						log.Println("[network] Error unmarshaling BackupMessage")
					}

				} else if event >= 4 && event <= 10 {
					var orderMessage = ElevatorOrderMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &orderMessage); err == nil { //unmarshal into correct message struct
						//log.Printf("[network] Received a orderMessage from %v with event %v \n", orderMessage.SenderIP, EventType[orderMessage.Event])
						receiveOrderChannel <- orderMessage
					} else {
						log.Println("[network] Error unmarshaling orderMessage")
					}

				}
			}

		}
	}
}

// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg#%d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		switch elemType.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
				elemType.String(), i+1))
		case reflect.Map:
			if elemType.Key().Kind() != reflect.String {
				panic(fmt.Sprintf(
					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
					elemType.String(), i+1))
			}
		}
	}
}
