package proxy

import (
	"net"
	"strconv"

	"uproxy/internal/request"

	log "github.com/sirupsen/logrus"
)

const protoHTTP = "HTTP"

func (p *Proxy) handleHttp(lConn *net.TCPConn, req *request.HTTPRequest, ip string) {
	req.SanitizeRequest()

	var port = 80
	var err error
	if req.Port() != "" {
		port, err = strconv.Atoi(req.Port())
		if err != nil {
			log.Debugf("[HTTP] Unable to parse port for %s. Aborting request due to invalid port format.", req.Domain())
		}
	}

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		lConn.Close()
		log.Debugf("[HTTP] Failed to connect to %s on port %d: %s", req.Domain(), port, err)
		return
	}

	log.Debugf("[HTTP] Established a connection to the server from %s to %s", rConn.LocalAddr(), req.Domain())

	go Serve(rConn, lConn, req.Domain(), lConn.RemoteAddr().String())

	_, err = rConn.Write(req.RawData())
	if err != nil {
		log.Debugf("[HTTP] Error while sending request to %s: %s. Ensure the server is reachable.", req.Domain(), err)
		return
	}

	log.Debugf("[HTTP] Successfully sent the request to %s", req.Domain())

	go Serve(lConn, rConn, lConn.RemoteAddr().String(), req.Domain())
}
