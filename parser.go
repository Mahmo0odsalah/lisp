package main

//go:generate enumgen

// RFC 3261

import (
	"errors"
	"strings"
)

// SIP MESSAGE structure:
// generic-message  =  start-line
// 										*message-header
// 										CRLF
// 										[ message-body ]
// start-line       =  Request-Line / Status-Line

// TODO: Handle header-values spanning multiple lines

type SIPMessageType int32 //enums:enum
const (
	SIPRequest SIPMessageType = iota
	SIPResponse
)

type Header struct {
	Name string
	Value string
}

type RequestLine struct {
	Method string // TODO: Revise type, enum?
	URI string
	SipVersion string
}

type StatusLine struct {
	StatusCode string // TODO: Revise type, enum?
	ReasonPhrase string
	SipVersion string
}

type SIPMessage struct {
	Mtype 	SIPMessageType
	RequestLine RequestLine
	StatusLine StatusLine
	// Why not a map? RFC3261 Section: 7.3.1
	Headers   []Header // TODO: Validate that needed headers are there
	Body      string
}

func (msg SIPMessage) FindHeaderByName(name string) (string, error) {
	for _, header := range(msg.Headers){
		if (header.Name == name){
			return header.Value, nil
		}
	}
	return "", errors.New("No header found with the provided name")
}

func Parse(packet []byte) (m SIPMessage){ //TODO: Handle errors
	// TODO: Look into optimizing this, sometimes we only need to read the startLine (or some headers) before making a decision, no need to parse the whole packet if this is taxing, examine if it's taxing first

	// TODO: Look into these headers: (Route, Record-Route, Proxy-Require, Max-Forwards, and Proxy-Authorization)

	strMsg := string(packet)
	lines := strings.Split(strMsg, "\n")

	startLine := strings.Split(lines[0], " ")
	startLine[2] = strings.TrimSpace(startLine[2])

	no_of_headers := len(lines) - 3
	m.Headers = make([]Header, no_of_headers)
	if startLine[0] == "SIP/2.0"	 { // Response
		m.Mtype = SIPResponse
		m.StatusLine = StatusLine{
			StatusCode: startLine[1],
			ReasonPhrase: startLine[2],
			SipVersion: startLine[0],
		}
	}	else { // Request
		// TODO: Validate that it's a request
		m.Mtype = SIPRequest
		m.RequestLine = RequestLine{
			Method: startLine[0],
			URI: startLine[1],
			SipVersion: startLine[2],
		}
	}
	for i, line := range lines[1:] {
		if line == "\r" { // Marks Headers ending
			break
		}

		hd := strings.Split(strings.TrimSpace(line), ": ")
		m.Headers[i] = Header{
			Name: hd[0],
			Value: hd[1],
		}
		
	}

	m.Body = lines[len(lines)-1]

	return m
}
