package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	result := ""
	prompt := &survey.Select{
		Message: "Select Mode:",
		Options: []string{"Diagnostics - Simple", "Diagnostics - Full", "Diagnostics - Custom", "Exit"},
	}

	if err := survey.AskOne(prompt, &result); err != nil {
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
	f, err := os.Create("diagnostics.txt")
	logger := io.MultiWriter(os.Stdout, f)
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
		}
	}(f)

	getInterfaceAddresses(logger)

	dns, err := dnsClientConfig()
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
	}

	_, _ = fmt.Fprintf(logger, "DNS Server(s): %s\n\n", strings.Join(dns.Servers, ", "))

	dns.Servers = append([]string{"system"}, dns.Servers...)
	if len(dns.Servers) <= 2 {
		dns.Servers = []string{"system"}
	}

	switch result {
	case diagSimple:
		simple(systemResolver, logger)
	case diagFull:
		simple(systemResolver, logger)
		full(dns.Servers, logger, f)
	case diagCustom:
		custom(dns.Servers, logger)
	}

	main()
}

func simple(servers []string, logger io.Writer) {
	_, _ = fmt.Fprintf(logger, "Looking up Steam diagnostics address...\n")
	lookupHostnames(testHostname, nil, 6, servers, logger, nil, false)
}

func full(servers []string, logger io.Writer, logfile *os.File) {
	for _, cdn := range CDNs {
		hostnames := parseCDN(cdn.Name, cdn.File, logger)
		lookupHostnames("", hostnames, 1, servers, logger, logfile, true)
	}
}

func custom(servers []string, logger io.Writer) {
	var result []string

	prompt := &survey.MultiSelect{
		Message: "Select CDN(s):",
		Options: []string{"ArenaNet", "Blizzard", "Battle State Games", "City of Heroes", "Daybreak Games",
			"Epic Games", "Frontier", "Neverwinter", "Nexus Mods", "Nintendo", "Origin", "Path of Exile",
			"RenegadeX", "Riot Games", "Rockstar Games", "Sony", "SQUARE ENIX", "Steam", "The Elder Scrolls Online",
			"UPlay", "Warframe", "WARGAMING", "Windows Updates", "Xbox Live"},
		PageSize: 24,
	}

	err := survey.AskOne(prompt, &result)
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: prompt failed %w", err))
		return
	}

	for _, cdn := range result {
		for _, cdns := range CDNs {
			if cdn == cdns.Name {
				hostnames := parseCDN(cdn, cdns.File, logger)
				lookupHostnames("", hostnames, 1, servers, logger, nil, false)
			} else {
				continue
			}
		}
	}
}
