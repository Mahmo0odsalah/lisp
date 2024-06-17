package main

import (
	"fmt"
	"net"
	"time"
)

var PublicIP string
var PrivateIP string
var UpstreamIP net.IP
var UpstreamPort int
var ProxyPort int

func main(){
	var err error

	PublicIP, err = GetPublicIP()
	if err != nil{
		panic(fmt.Sprintf("Unable to retrieve public IP: %v", err))
	}

	PrivateIP, err = GetPrivateIP()
	if err != nil {
		panic(fmt.Sprintf("Unable to retrieve private IP: %v", err))
	}
	
	ProxyPort = 5060
	// TODO: Support TCP receiving and sending (sometimes depends on Via header)
  conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: ProxyPort})

	UpstreamIP = net.IPv4(127,0,0,1)
	UpstreamPort = 5061

  if err != nil { // TODO: Properly handle errors
    fmt.Println(err)
  }

	// The concurrency model here aims to achieve 3 things:
	// 1. Network operations (send & recv) are always serial. To avoid any race conditions corruputing packets.
	// 2. Order is important. Packets need to be proxied with the order they are received at, to avoid messing up one party's SIP flow.
	// 3. Anything not related to network operations should be concurrent

    
  b := make([]byte, 65_535)
	channels := make([]chan *ProxyResult, 0, 100_000)

	go func(){ // The goroutine responsible for sending packets
		i := 0
		for {
			var chann chan *ProxyResult

			for chann == nil {
				if len(channels) > i {
					chann = channels[i]
				}
				time.Sleep(1 * time.Millisecond)
			}
			res :=  <- chann

			if res != nil {
				_, _, err = conn.WriteMsgUDP(res.newPacket, []byte{}, res.targetAddr)
				
				if err != nil { // TODO: Properly handle errors
					fmt.Println(err)
				}

			}
			i += 1

			// every 1000 packets, resize the slice to avoid memory bloating
			if (i > 1000){
				channels = channels[i:]
				i = 0
			}
		}
	}()

	for {
		n, _, _, _, err := conn.ReadMsgUDP(b, []byte{})
		
		if err != nil { // TODO: Properly handle error
      fmt.Println(err)
    } else { // No point in handling partially receive packets in case of errors, the SIP protocol handles re-transmitting important information if it's never ACK'd
    	packet := make([]byte, n)
			copy(packet, b[0:n])
			ch := make(chan *ProxyResult)
			channels = append(channels, ch)
			go Proxy(ch, packet)
		}

	
  }
}
