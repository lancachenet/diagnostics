package main

import (
	"github.com/miekg/dns"
)

func dnsClientConfig() (*dns.ClientConfig, error) {
	return dns.ClientConfigFromFile(resolvConf)
}
