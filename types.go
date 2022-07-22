package main

import "net"

type CDN struct {
	Name string
	File string
}

type Lookup struct {
	Hostname    string
	Address     []net.IP
	ContainerID string
}
