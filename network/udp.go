package network

import (
	"log"
	"net"
	"strconv"

	. "../utilities"
)

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 4*1024
const localListenPort = 44044 // Port for inncomming udp messages
const broadcastSendPort = 44033 // Port for outgoing udp message

type UDPMessage struct {
	Raddr   string 			// MsgMessageChannel or MsgBackupChannel
	Data   []byte
	Length int			// length of received data, empt when sending
}

const(
	MsgMessageChannel = iota
	MsgBackupChannel
)

var broadcastAddr *net.UDPAddr
var localAddr			 *net.UDPAddr
var localIP string

func InitUDP(
	udpSendDatagramChannel <-chan UDPMessage,
	udpReceiveDatagramChannel chan<- UDPMessage) (localIP string, err error) {

	broadcastAddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255"+":"+strconv.Itoa(broadcastSendPort))
	CheckError("ERROR [udp]: Failed to resolve remote addr", err)

	localAddr, err = net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(localListenPort))
	CheckError("ERROR [udp]: Failed to resolve broadcastlocalListenConnPort: ", err)

	// Broadcast broadcastlocalListenConnConnection
	broadcastListenConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	CheckError("[udp] ERROR DialUDP failed", err)

	// Local localListenConning connection
	localListenConn, err := net.ListenUDP("udp4", localAddr)
	CheckError("[udp] Failed to create local listen connection", err)

	// Get local IP address
	localIP, err = resolveLocalIP(broadcastAddr)
	CheckError("ERROR [udp]: Failed to get local addr", err)
	printDebug("[udp] LocalIP:" + localIP)





	go udpTransmit(localListenConn, udpSendDatagramChannel)
	go udpReceive(broadcastListenConn, udpReceiveDatagramChannel)

	return localIP, nil
}

func resolveLocalIP(broadcastAddr *net.UDPAddr) (string, error) {
	tempConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		log.Println("[udo] resolveLocalIP, no internet connection", err)
		return "", err
	} else {
		defer tempConn.Close()
	}
	localIP = tempConn.LocalAddr().String()
	//strings.Split(conn.LocalAddr().String(), ":")[0]
	return localIP, nil

}



func udpTransmit(conn *net.UDPConn, udpSendDatagramChannel <-chan UDPMessage) {
	defer conn.Close()
	for {
		select {
		case message := <-udpSendDatagramChannel:
			if debug{
				//log.Println("[udp] UDP Send: \t", string(message.Data))
			}
			n, err := conn.WriteToUDP(message.Data, broadcastAddr)
			if (err != nil || n < 0) && debug {
				log.Println("[udp] Sending UDP broadcast failed", err)
			} else {
				//printDebug("[udp] UDP Sent number of bytes:" + strconv.Itoa(n) )
			}
		}
	}
}

func udpReceive(conn *net.UDPConn, udpReceiveDatagramChannel chan<- UDPMessage) {
	defer conn.Close()
	for {
		if debug {
			log.Printf("[udp] UDPConnectionReader:\t Waiting on data from UDPConn %s\n", localIP)
		}
		buf := make([]byte, messageSize)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil || n < 0 || n > messageSize {
			log.Println("[udp]  Error in ReadFromUDP:", err)
		} else {
			if debug {
				log.Printf("[udp] udpReceive Received packet from: ", raddr.String())
				log.Printf("[udp] udpReceive: \t", string(buf[:]))
			}
			log.Println("from udpReceive")

		udpReceiveDatagramChannel <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}
		}
	}
}
