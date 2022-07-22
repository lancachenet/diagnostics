package main

import "net"

type CDN struct {
	Name string
	File string
}

type Lookup struct {
	Resolver    string
	Hostname    string
	Address     []net.IP
	ContainerID string
	Time        string
}
