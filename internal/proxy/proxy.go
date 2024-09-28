package proxy

import (
	"context"
	"net"
	"os"
	"strconv"

	"uproxy/internal/config"
	"uproxy/internal/dns"
	"uproxy/internal/request"

	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	addr     string
	port     int
	resolver *dns.DNS
}

func New(config *config.Config) *Proxy {
	return &Proxy{
		addr:     *config.Addr,
		port:     *config.Port,
		resolver: dns.NewDns(config),
	}
}

func (p *Proxy) Start(ctx context.Context) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(p.addr), Port: p.port})
	if err != nil {
		log.Fatal("proxy error creating listener: ", err)
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down proxy server...")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Println("proxy error accepting connection: ", err)
				continue
			}

			go p.handleConnection(ctx, conn)
		}
	}
}

func (p *Proxy) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	req, err := request.ParseHTTPRequest(conn)
	if err != nil {
		log.Println("proxy error while parsing request: ", err)
		return
	}

	log.Println("proxy request from ", conn.RemoteAddr(), "\n\n", string(req.RawData()))

	if !req.IsSupportedMethod() {
		log.Println("proxy unsupported method: ", req.Method())
		return
	}

	ip, err := p.resolver.ResolveHost(ctx, req.Domain())
	if err != nil {
		log.Debug("proxy error while dns lookup: ", req.Domain(), " ", err)
		conn.Write([]byte(req.Version() + " 502 Bad Gateway\r\n\r\n"))
		return
	}

	if p.isLoopedRequest(req, net.ParseIP(ip)) {
		log.Error("[PROXY] looped request has been detected. aborting.")
		return
	}

	p.handleRequest(conn.(*net.TCPConn), req, ip)
}

func (p *Proxy) handleRequest(conn *net.TCPConn, req *request.HTTPRequest, ip string) {
	if req.IsMethodConnect() {
		p.handleHttps(conn, req, ip)
	} else {
		p.handleHttp(conn, req, ip)
	}
}

func (p *Proxy) isLoopedRequest(req *request.HTTPRequest, ip net.IP) bool {
	if req.Port() != strconv.Itoa(p.port) {
		return false
	}

	if ip.IsLoopback() {
		return true
	}

	addr, err := net.InterfaceAddrs()
	if err != nil {
		log.Error("[PROXY] error while getting addresses of our network interfaces: ", err)
		return false
	}

	for _, addr := range addr {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.Equal(ip) {
				return true
			}
		}
	}

	return false
}
