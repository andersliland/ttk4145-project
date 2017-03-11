package network

import (
	"log"
	"net"
	"strconv"
	"strings"

	. "../utilities"
)

const debugUDP = false

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 4 * 1024
const broadcastSendPort = 44077 // SendTo and ListenFrom port

type UDPMessage struct {
	Raddr  string // MsgMessageChannel or MsgBackupChannel
	Data   []byte
	Length int // length of received data, empt when sending
}

var broadcastAddr *net.UDPAddr
var localAddr *net.UDPAddr
var localIP string

func InitUDP(
	udpSendDatagramChannel <-chan UDPMessage,
	udpReceiveDatagramChannel chan<- UDPMessage) (localIP string, err error) {

	broadcastAddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255"+":"+strconv.Itoa(broadcastSendPort))
	CheckError("ERROR [udp]: Failed to resolve remote addr", err)

	localAddr, err = net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(broadcastSendPort))
	CheckError("ERROR [udp]: Failed to resolve broadcastlocalListenConnPort: ", err)

	// Get local IP address
	localIP, err = resolveLocalIP(broadcastAddr)
	CheckError("ERROR [udp]: Failed to get local addr", err)
	if debugUDP {
		log.Println("[udp] LocalIP:" + localIP)

	}

	// Broadcast broadcastlocalListenConnConnection
	broadcastSendConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	CheckError("[udp] ERROR DialUDP failed", err)

	// Local localListenConning connection
	listen, err := net.ListenUDP("udp4", localAddr)
	CheckError("[udp] Failed to create local listen connection", err)

	go udpTransmit(broadcastSendConn, udpSendDatagramChannel)
	go udpReceive(listen, udpReceiveDatagramChannel)

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
	localIP = strings.Split(tempConn.LocalAddr().String(), ":")[0]
	return localIP, nil
}

func udpTransmit(conn *net.UDPConn, udpSendDatagramChannel <-chan UDPMessage) {
	defer conn.Close()
	for {
		select {
		case message := <-udpSendDatagramChannel:
			if debugUDP {
				//log.Println("[udp] UDP Send: \t", string(message.Data))
			}
			n, err := conn.Write(message.Data)
			if (err != nil || n < 0) && debugUDP {
				log.Println("[udp] Sending UDP broadcast failed", err)
			} else {
				if debugUDP {
					log.Println("[udp] UDP Sent number of bytes:" + strconv.Itoa(n))
				}
			}
		}
	}
}

func udpReceive(conn *net.UDPConn, udpReceiveDatagramChannel chan<- UDPMessage) {

	bconn_rcv_ch := make(chan UDPMessage, 5)
	go udpConnectionReader(conn, bconn_rcv_ch)
	for {
		select {
		case f := <-bconn_rcv_ch:
			udpReceiveDatagramChannel <- f
		}
	}

}

func udpConnectionReader(conn *net.UDPConn, bconn_rcv_ch chan<- UDPMessage) {
	buf := make([]byte, messageSize)
	for {
		if debugUDP {
			log.Printf("[udp] UDPConnectionReader:\t Waiting on data from UDPConn %s\n", localIP)
		}
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil || n < 0 || n > messageSize {
			log.Println("[udp]  Error in ReadFromUDP:", err)
		} else {
			if debugUDP {
				log.Printf("[udp] udpReceive Received packet from: %v ", raddr.String())
				log.Printf("[udp] udpReceive: \t %v", string(buf[:]))
			}
			bconn_rcv_ch <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}
		}
	}

}

func printUDP(s string) {
	if debugUDP {
		log.Println("[udp]", s)
	}
}
