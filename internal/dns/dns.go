package dns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"uproxy/internal/config"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Resolver interface {
	Resolve(ctx context.Context, host string, qTypes []uint16) ([]net.IPAddr, error)
}

type DNS struct {
	host     string
	port     string
	resolver Resolver
}

func NewDns(config *config.Config) *DNS {
	return &DNS{
		host:     *config.DnsAddr,
		port:     strconv.Itoa(*config.DnsPort),
		resolver: NewSystemResolver(),
	}
}

func (d *DNS) ResolveHost(ctx context.Context, host string) (string, error) {
	if ip, err := parseIpAddr(host); err == nil {
		return ip.String(), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := time.Now()

	addrs, err := d.resolver.Resolve(ctx, host, []uint16{dns.TypeAAAA, dns.TypeA})
	if err != nil {
		return "", fmt.Errorf("resolver error: %w", err)
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("could not resolve %s", host)
	}

	log.Debug("resolved %s from %s in %d ms", addrs[0].String(), host, time.Since(startTime).Milliseconds())

	return addrs[0].String(), nil
}

func parseIpAddr(addr string) (*net.IPAddr, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("%s is not an ip address", addr)
	}

	return &net.IPAddr{IP: ip}, nil
}
