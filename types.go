package main

type CDN struct {
	Name string
	File string
}

type Lookup struct {
	Resolver    string
	Hostname    string
	Address     []string
	ContainerID string
	Time        string
}
