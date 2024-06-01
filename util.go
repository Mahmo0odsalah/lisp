package main

import (
	"io"
	"net"
	"net/http"
)

func IsPublicIP(IP net.IP) bool {
		// TODO: Add functionality to configure custom local IP subnet range
    if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
        return false
    }
    if ip4 := IP.To4(); ip4 != nil {
        switch {
        case ip4[0] == 10:
            return false
        case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
            return false
        case ip4[0] == 192 && ip4[1] == 168:
            return false
        default:
            return true
        }
    }
    return false
}

// Get preferred outbound ip of this machine
func GetPrivateIP() (string, error) {
    conn, err := net.Dial("udp", "8.8.8.8:80")

		if err != nil {
			return "", err
    }

		defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String(), nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://ifconfig.me/ip")

	if (err != nil){
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}