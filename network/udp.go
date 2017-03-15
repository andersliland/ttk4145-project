package network

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	. "../utilities"
)

const debugUDP = false

// Maximum allowed UDP datagram size in bytes: 65,507 (imposed by the IPv4 protocol)
const messageSize = 4 * 1024
const broadcastSendPort = 44077

type UDPMessage struct {
	Raddr  string
	Data   []byte
	Length int // length of received data, empty when sending
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
		return "", err
	}

	localAddr, err = net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(broadcastSendPort))
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to local remote adress. Are you connected to the internet?", ColorNeutral)
		return "", err
	}

	// Get local IP address
	localIP, err = resolveLocalIP(broadcastAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to get own IP adress. Are you connected to the internet??", ColorNeutral)
		return "", err
	}

	// Broadcast broadcastlocalListenConnConnection
	broadcastSendConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to dial UDP broadcast connection", ColorNeutral)
		return "", err
	}

	// Local localListenConning connection
	listen, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udp]\t\t Failed to dial UDP listen connection", ColorNeutral)
		return "", err
	}

	go udpTransmit(broadcastSendConn, udpSendDatagramChannel)
	go udpReceive(listen, udpReceiveDatagramChannel)

	return localIP, nil
}

func resolveLocalIP(broadcastAddr *net.UDPAddr) (string, error) {
	tempConn, err := net.DialUDP("udp4", nil, broadcastAddr)
	if err != nil {
		fmt.Print(ColorRed)
		log.Println("[udo]\t\t Failed to get own IP adress. Are you connected to the internet?", ColorNeutral)
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
				printUDP("UDP Sent number of bytes: " + strconv.Itoa(n))
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

		buf := make([]byte, messageSize) // Moved inside for loop to clear buffer between each message

		if debugUDP {
			log.Println("[udp]\t\t UDPConnectionReader:\t Waiting on data from UDPConn " + localIP)
		}
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil || n < 0 || n > messageSize {
			log.Println("[udp]\t\t  Error in ReadFromUDP:", err)
		} else {
			printUDP("[udp]\t\t Received UDP packet from: %v " + raddr.String())
			printUDP("[udp]\t\t With data " + string(buf[:]))
			bconn_rcv_ch <- UDPMessage{Raddr: raddr.String(), Data: buf[:n], Length: n}
		}
	}
}

func printUDP(s string) {
	if debugUDP {
		log.Println("[udp]\t\t", s)
	}
}
