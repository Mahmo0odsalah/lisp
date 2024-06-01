package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func main(){
  addr := net.UDPAddr{IP: nil,Port: 5060}
	// TODO: Support TCP receiving and sending (sometimes depends on Via header)
  conn, err := net.ListenUDP("udp", &addr)

	upstreamIP := net.IPv4(127,0,0,1)
	upstreamPort := 5061

  if err != nil { // TODO: Properly handle errors
    fmt.Println(err)
  }
  
	if err != nil { // TODO: Properly handle errors
		fmt.Println(err)
	}

  b := make([]byte, 100000) // TODO: Change to the theoritical max packet size (or SIP packet size if exists)

  for {
    n, _, _, src, err := conn.ReadMsgUDP(b, []byte{})
		
		// TODO: GOroutine-ify
    fmt.Printf("Read %d bytes from %v\n", n, src)

		// TODO: Do we want to parse the partially received packet in case of an error?
    packet := b[0:n]

    
    if err != nil { // TODO: Properly handle error
      fmt.Println(err)
    }
		
		p := Parse(packet)
		
		// Step 1: Determine if request or response, if response goto step 5
		// Step 2: If this is an INVITE, respond with 100 trying
		// Step 3: Add A VIA header before the original VIA header that indicates the proxy's endpoint, RFC3261 P:13
		// Step 4: Proxy the request to the target
		// Step 5: This is a response, remove the proxy's VIA from the packet
		// Step 6: Determine the target using the now top VIA header
		// Step 7: Proxy

		var remoteIP net.IP
		var remotePort int

		if (p.Mtype == SIPRequest){
			if (p.RequestLine.Method == "INVITE"){
				// TODO: Respond with 100 Trying
			}
			// Proxying to upstream
			remoteIP = upstreamIP
			remotePort = upstreamPort
			// TODO: Add VIA header for proxy
		} else {
			if (p.StatusLine.StatusCode == "100"){
				continue // Proxy already responds with 100 Trying, no need to proxy 100s
			} 

			// TODO: Remove the proxy's VIA header
			via, _ := p.FindHeaderByName("Via")
			// SIP/2.0/UDP 192.168.0.222:5062;rport;branch=z9hG4bKPjd87fd14a-5db8-4b66-a50b-c28bee9cc49c
			transport := strings.Split(via, ";")[0]
			ip, port, _ := net.SplitHostPort(strings.Split(transport, " ")[1])
			remoteIP = net.ParseIP(ip)
			remotePort, _ = strconv.Atoi(port)
		}
		targetAddr := net.UDPAddr{ IP: remoteIP, Port: remotePort }
		fmt.Printf("Proxying %v to %v\n", p, targetAddr)
		_, _, err = conn.WriteMsgUDP(packet, []byte{}, &targetAddr)

		if err != nil { // TODO: Properly handle error
      fmt.Println(err)
    }
  }
}
