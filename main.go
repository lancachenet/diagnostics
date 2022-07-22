package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func main() {
	prompt := promptui.Select{
		Label: "Select Mode",
		Items: []string{"Diagnostics - Simple", "Diagnostics - Full", "Diagnostics - Custom", "Exit"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Print(fmt.Errorf("error: prompt failed %w", err))
		return
	}

	if strings.Contains(result, "Diagnostics") {
		diagnostics(result)
	} else if result == "Exit" {
		os.Exit(0)
	}
}

func diagnostics(result string) {
	getInterfaceAddresses()

	dns, err := dnsClientConfig()
	if err != nil {
		fmt.Print(fmt.Errorf("error: %w", err))
	}

	fmt.Printf("DNS Server(s): %s\n\n", strings.Join(dns.Servers, ", "))

	dns.Servers = append([]string{"system"}, dns.Servers...)
	if len(dns.Servers) <= 2 {
		dns.Servers = []string{"system"}
	}

	switch result {
	case diagSimple:
		simple(systemResolver)
	case diagFull:
		simple(systemResolver)
		full(dns.Servers)
	case diagCustom:
		custom(dns.Servers)
	}

	main()
}

func simple(servers []string) {
	fmt.Printf("Looking up Steam diagnostics address...\n")
	lookupHostnames(testHostname, nil, 6, servers)
}

func full(servers []string) {
	for _, cdn := range CDNs {
		hostnames := parseCDN(cdn.Name, cdn.File)
		lookupHostnames("", hostnames, 1, servers)
	}
}

func custom(servers []string) {
selection:
	file := ""
	prompt := promptui.Select{
		Label: "Select CDN",
		Items: []string{"Back", "ArenaNet", "Blizzard", "Battle State Games", "City of Heroes", "Daybreak Games",
			"Epic Games", "Frontier", "Neverwinter", "Nexus Mods", "Nintendo", "Origin", "Path of Exile",
			"RenegadeX", "Riot Games", "Rockstar Games", "Sony", "SQUARE ENIX", "Steam", "The Elder Scrolls Online",
			"UPlay", "Warframe", "WARGAMING", "Windows Updates", "Xbox Live", "Exit"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Print(fmt.Errorf("error: prompt failed %w", err))
		return
	}

	switch result {
	case "Back":
		main()
	case "Exit":
		os.Exit(0)
	default:
		for _, cdn := range CDNs {
			if cdn.Name == result {
				file = cdn.File
			} else {
				continue
			}
			fmt.Printf("No matching CDN file found: %s\n", file)
		}
	}

	hostnames := parseCDN(result, file)
	lookupHostnames("", hostnames, 1, servers)

	goto selection
}
