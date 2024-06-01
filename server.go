package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func main(){
	publicIp, err := GetPublicIP()
	if (err != nil){
		panic(fmt.Sprintf("Unable to retrieve public IP: %v", err))
	}
	privateIp, err := GetPrivateIP()
	if (err != nil) {
		panic(fmt.Sprintf("Unable to retrieve private IP: %v", err))
	}
	
	proxyPort := 5060
  addr := net.UDPAddr{IP: nil, Port: proxyPort}
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

		// TODO: Check that Max-Forwards didn't reach 0
		// TODO: Decrement Max-Forwards

		if (p.Mtype == SIPRequest){
			if (p.RequestLine.Method == "INVITE"){
				// TODO: Respond with 100 Trying
			}
			// Proxying to upstream
			remoteIP = upstreamIP
			remotePort = upstreamPort
			fmt.Println("Before:")
			fmt.Println(p)

			// Prefer Proxy's private IP when possible
			var proxyIp string
			if (IsPublicIP(upstreamIP)){
				proxyIp = publicIp
			} else {
				proxyIp = privateIp
			}
			via, _ := p.FindHeaderByName("Via")
			branch := strings.Split(via, ";")[2]
			proxyVia := Header{ Name: "Via", Value: fmt.Sprintf("SIP/2.0/UDP %s:%d;rport;%s",proxyIp, proxyPort, branch)}
			newHeaders := make([]Header, len(p.Headers) + 1)
			newHeaders[0] = proxyVia
			copy(newHeaders[1:], p.Headers)
			p.Headers = newHeaders
		} else {
			if (p.StatusLine.StatusCode == "100"){
				continue // Proxy already responds with 100 Trying, no need to proxy 100s
			} 

			newHeaders := make([]Header, len(p.Headers) - 1)
			viaRemoved := false
			for _, hd := range newHeaders {
				if (!viaRemoved && hd.Name == "Via") {
					continue
				}
				newHeaders = append(newHeaders, hd)
			}

			via, _ := p.FindHeaderByName("Via")
			// SIP/2.0/UDP 192.168.0.222:5062;rport;branch=z9hG4bKPjd87fd14a-5db8-4b66-a50b-c28bee9cc49c
			transport := strings.Split(via, ";")[0]
			ip, port, _ := net.SplitHostPort(strings.Split(transport, " ")[1])
			remoteIP = net.ParseIP(ip)
			remotePort, _ = strconv.Atoi(port)
		}
		targetAddr := net.UDPAddr{ IP: remoteIP, Port: remotePort }
		fmt.Printf("Proxying %v to %v\n", p, targetAddr)

		// TODO: Send the new packet
		_, _, err = conn.WriteMsgUDP(packet, []byte{}, &targetAddr)

		if err != nil { // TODO: Properly handle error
      fmt.Println(err)
    }
  }
}
