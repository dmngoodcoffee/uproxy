package config

import (
	"flag"
)

type Config struct {
	Addr    *string
	Port    *int
	DnsAddr *string
	DnsPort *int
}

func New() *Config {
	conf := new(Config)
	conf.Addr = flag.String("addr", "127.0.0.1", "listen address")
	conf.Port = flag.Int("port", 8080, "port")
	conf.DnsAddr = flag.String("dns-addr", "8.8.8.8", "dns address")
	conf.DnsPort = flag.Int("dns-port", 53, "port number for dns")

	flag.Parse()

	return conf
}
