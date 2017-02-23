package network

import (
	"encoding/json"
	"log"

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
			networkPack, err := json.Marshal(message)
			//log.Println("[network] Message to be sendt", message)
			if err != nil {
				log.Println("ERROR [network]: sendMessage marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Raddr: "32", Data: networkPack} // UDPMessage
				//log.Println("[network] Sent a Message with event", EventType[message.Event])

			}
		case message := <-sendBackupChannel:
			networkPack, err := json.Marshal(message)
			if err != nil {
				log.Println("ERROR [network]: sendBackup marshal failed", err)
			} else {
				udpSendDatagramChannel <- UDPMessage{Raddr: "255.255.255.255", Data: networkPack} // UDPMessage
				//log.Println("[network] Sent a BackupMessage with event", EventType[message.Event])
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
				event := int(m["Event"].(float64)) // type assertion, float64 because
				//log.Printf("event %T, %v", event, event)
				if event <= 3 && event >= 0 {
					var backupMessage = ElevatorBackupMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &backupMessage); err == nil { //unmarshal into correct message struct
						// TODO: make check to validate message
						receiveBackupChannel <- backupMessage

					} else {
						log.Println("[network] Error unmarshaling BackupMessage")
					}

				} else if event >= 4 && event <= 10 {
					var orderMessage = ElevatorOrderMessage{}
					if err := json.Unmarshal(msg.Data[:msg.Length], &orderMessage); err == nil {
						// TODO: validate unmarshal
						receiveOrderChannel <- orderMessage
					} else {
						log.Println("[network] Error unmarshaling OrderMessage")
					}
				}

				//fmt.Println("m ", m)
				//fmt.Println("event ", event)
			}

		}
	}
}
