package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ProxyResult struct {
	targetAddr *net.UDPAddr
	newPacket []byte
}

// Step 1: Determine if request or response, if response goto step 5
// Step 2: If this is an INVITE, respond with 100 trying
// Step 3: Add A VIA header before the original VIA header that indicates the proxy's endpoint, RFC3261 P:13
// Step 4: Proxy the request to the target, go to end.
// Step 5: This is a response, remove the proxy's VIA from the packet
// Step 6: Determine the target using the now top VIA header
// Step 7: Proxy

func Proxy(result chan<- *ProxyResult, packet []byte){
		p := Parse(packet)

		var remoteIP net.IP
		var remotePort int

		// TODO: Check that Max-Forwards didn't reach 0
		// TODO: Decrement Max-Forwards
		if p.Mtype == SIPRequest {
			if p.RequestLine.Method == "INVITE" {
				// TODO: Respond with 100 Trying
			}
			// Proxying to upstream
			remoteIP = UpstreamIP
			remotePort = UpstreamPort

			// Prefer Proxy's private IP when possible
			var proxyIp string
			if IsPublicIP(UpstreamIP) {
				proxyIp = PublicIP
			} else {
				proxyIp = PrivateIP
			}
			via, _ := p.FindHeaderByName("Via")
			branch := strings.Split(via, ";")[2]
			proxyVia := Header{ Name: "Via", Value: fmt.Sprintf("SIP/2.0/UDP %s:%d;rport;%s",proxyIp, ProxyPort, branch)}
			newHeaders := make([]Header, len(p.Headers) + 1)
			newHeaders[0] = proxyVia
			copy(newHeaders[1:], p.Headers)
			p.Headers = newHeaders
		} else {
			if p.StatusLine.StatusCode == "100" {
				// Proxy already responds with 100 Trying, no need to proxy 100s
				close(result)
				return
			} 

			newHeaders := make([]Header, 0, len(p.Headers) - 1)
			viaRemoved := false
			for _, hd := range p.Headers {
				if !viaRemoved && hd.Name == "Via" {
					viaRemoved = true
					continue
				}
				newHeaders = append(newHeaders, hd)
			}
			p.Headers = newHeaders

			via, err := p.FindHeaderByName("Via")

			if err != nil {
				fmt.Println(err)
			}
			// SIP/2.0/UDP 192.168.0.222:5062;rport;branch=z9hG4bKPjd87fd14a-5db8-4b66-a50b-c28bee9cc49c
			transport := strings.Split(via, ";")[0]
			ip, port, _ := net.SplitHostPort(strings.Split(transport, " ")[1])
			remoteIP = net.ParseIP(ip)
			remotePort, _ = strconv.Atoi(port)
		}
		targetAddr := net.UDPAddr{ IP: remoteIP, Port: remotePort }

		newPacket := []byte (p.String())

		result <- &ProxyResult{
			targetAddr: &targetAddr,
			newPacket: newPacket,
		}
}