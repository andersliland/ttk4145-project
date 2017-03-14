package network

import (
	"fmt"
	"log"
	"net"
	"os"
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
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to resolve remote adress. Are you connected to the internet?", ColorNeutral)
		log.Println("[udp]\t\t ", err)
	}

	localAddr, err = net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(broadcastSendPort))
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to local remote adress. Are you connected to the internet?", ColorNeutral)
		log.Println("[udp]\t\t ", err)
	}

	// Get local IP address
	localIP, err = resolveLocalIP(broadcastAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to get own IP adress. Does your computer have a network card?", ColorNeutral)
		log.Println("[udp]\t\t ", err)
	}

	// Broadcast broadcastlocalListenConnConnection
	broadcastSendConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to dial UDP broadcast connection", ColorNeutral)
		log.Println("[udp]\t\t ", err)
		os.Exit(1)

	}

	// Local localListenConning connection
	listen, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to dial UDP listen connection", ColorNeutral)
		log.Println("[udp]\t\t ", err)
		os.Exit(1)
	}

	go udpTransmit(broadcastSendConn, udpSendDatagramChannel)
	go udpReceive(listen, udpReceiveDatagramChannel)

	return localIP, nil
}

func resolveLocalIP(broadcastAddr *net.UDPAddr) (string, error) {
	tempConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		log.Println("[udo]\t\t resolveLocalIP, no internet connection", err)
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
			n, err := conn.Write(message.Data)
			if (err != nil || n < 0) && debugUDP {
				log.Println("[udp]\t\t Sending UDP broadcast failed", err)
			} else {
				if debugUDP {
					log.Println("[udp]\t\t UDP Sent number of bytes:" + strconv.Itoa(n))
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
	for {
		buf := make([]byte, messageSize)

		if debugUDP {
			log.Printf("[udp]\t\t UDPConnectionReader:\t Waiting on data from UDPConn %s\n", localIP)
		}
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil || n < 0 || n > messageSize {
			log.Println("[udp]\t\t  Error in ReadFromUDP:", err)
		} else {
			if debugUDP {
				log.Printf("[udp]\t\t udpReceive Received packet from: %v ", raddr.String())
				log.Printf("[udp]\t\t udpReceive: \t %v", string(buf[:]))
			}
			bconn_rcv_ch <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}
		}
	}

}

func printUDP(s string) {
	if debugUDP {
		log.Println("[udp]\t", s)
	}
}
