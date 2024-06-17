package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"gopkg.in/matryer/try.v1"
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
		panic(fmt.Sprintf("Error while retrieving public IP: %v", err))
	}

	PrivateIP, err = GetPrivateIP()
	if err != nil {
		panic(fmt.Sprintf("Error while retrieving private IP: %v", err))
	}
	
	ProxyPort = 5060
	// TODO: Support TCP receiving and sending (sometimes depends on Via header)
  conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: ProxyPort})

	if err != nil {
    panic(fmt.Sprintf("Couldn't listen on provied port: %s", err))
  }
	defer conn.Close()

	UpstreamIP = net.IPv4(127,0,0,1)
	UpstreamPort = 5061



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
				
				err := try.Do(func(attempt int) (bool, error) {
					_, _, err := conn.WriteMsgUDP(res.newPacket, []byte{}, res.targetAddr)
					maxAttempts := 4
					retry := attempt < maxAttempts 
					if err != nil {
						log.Printf("Error while attempting to send a packet to %v, attempt no: %d", res.targetAddr, attempt)
						if retry {
							time.Sleep(time.Duration((10^attempt)) * time.Millisecond) // Exponential backoff with a max of 1 second (10^(maxAttempts-1))
						}
					}
					return retry , err
				})

				if err != nil {
					log.Printf("Failed to send packet to %v, continuing...", res.targetAddr) // The proxy shouldn't exit if it failed to send a connection to a specific party, otherwise it would be sensitive to an external UA disconnecting.
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
		
		if err != nil {
      log.Println("Error while reading packets", err)
    } else { // No point in handling partially receive packets in case of errors, the SIP protocol handles re-transmitting important information if it's never ACK'd
    	packet := make([]byte, n)
			copy(packet, b[0:n])
			ch := make(chan *ProxyResult)
			channels = append(channels, ch)
			go Proxy(ch, packet)
		}
  }
}
