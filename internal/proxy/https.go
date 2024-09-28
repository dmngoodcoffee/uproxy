package proxy

import (
	"net"
	"strconv"

	"uproxy/internal/request"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) handleHttps(lConn *net.TCPConn, req *request.HTTPRequest, ip string) {
	var port = 443
	var err error
	if req.Port() != "" {
		port, err = strconv.Atoi(req.Port())
		if err != nil {
			log.Debugf("[HTTPS] Unable to parse port for %s. Aborting request due to invalid port format.", req.Domain())
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		lConn.Close()
		log.Debugf("[HTTPS] Failed to establish a connection to %s on port %d: %s", req.Domain(), port, err)
		return
	}

	log.Debugf("[HTTPS] Established a connection to the server from %s to %s", rConn.LocalAddr(), req.Domain())

	_, err = lConn.Write([]byte(req.Version() + " 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Debugf("[HTTPS] Error while sending '200 Connection Established' to the client: %s", err)
		return
	}

	log.Debugf("[HTTPS] Successfully sent connection established response to %s", lConn.RemoteAddr())

	m, err := request.ReadTLSMessage(lConn)
	if err != nil || !m.IsClientHello() {
		log.Debugf("[HTTPS] Error reading Client Hello from %s: %s", lConn.RemoteAddr().String(), err)
		// return
	}
	clientHello := m.Raw

	log.Debugf("[HTTPS] Received Client Hello message of %d bytes", len(clientHello))

	// Generate a go routine that reads from the server
	go Serve(rConn, lConn, req.Domain(), lConn.RemoteAddr().String())

	if _, err := rConn.Write(clientHello); err != nil {
		log.Debugf("[HTTPS] Error while writing Client Hello to %s: %s", req.Domain(), err)
		return
	}

	go Serve(lConn, rConn, lConn.RemoteAddr().String(), req.Domain())
}
