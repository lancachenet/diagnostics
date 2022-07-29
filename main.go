package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/miekg/dns"
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

	d, err := dnsClientConfig()
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
	}

	_, _ = fmt.Fprintf(logger, "DNS Server(s): %s\n\n", strings.Join(d.Servers, ", "))

	d.Servers = append([]string{"system"}, d.Servers...)
	if len(d.Servers) <= 2 {
		d.Servers = []string{"system"}
	}

	switch result {
	case diagSimple:
		simple(systemResolver, logger)
	case diagFull:
		simple(systemResolver, logger)
		full(d.Servers, logger, f)
	case diagCustom:
		custom(d.Servers, logger)
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
	var (
		options []string
		result  []string
	)

	for _, cdn := range CDNs {
		options = append(options, cdn.Name)
	}

	prompt := &survey.MultiSelect{
		Message:  "Select CDN(s):",
		Options:  options,
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

func getInterfaceAddresses(logger io.Writer) {
	interfaces, err := net.Interfaces()
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
		return
	}

	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
			continue
		}

		if strings.Contains(i.Flags.String(), up) {
			if !strings.Contains(strings.ToLower(i.Name), loopback) && !strings.Contains(i.Flags.String(), loopback) {
				_, _ = fmt.Fprintf(logger, "Interface: %s\n", i.Name)

				for _, a := range addresses {
					switch v := a.(type) {
					case *net.IPAddr:
						_, _ = fmt.Fprintf(logger, "IP Address: %s\n", v)

					case *net.IPNet:
						_, _ = fmt.Fprintf(logger, "IP Address: %s\n", v)
					}
				}
				_, _ = fmt.Fprintf(logger, "\n")
			}
		}
	}
}

func lookupHostnames(host string, hostnames []string, iterations int, servers []string, logger io.Writer, logfile *os.File, debug bool) {
	var (
		lookups, success, failed, deltas []Lookup
	)

	for _, resolver := range servers {
		success = nil
		failed = nil
		resolverMsg := "with system resolver"
		if resolver != "system" {
			resolverMsg = fmt.Sprintf("with resolver: %s", resolver)
		}

		for i := 0; i < iterations; i++ {
			if host != "" {
				s, f := processHostnames(host, resolver, logger)
				success = append(success, s...)
				failed = append(failed, f...)
			} else {
				for _, hostname := range hostnames {
					s, f := processHostnames(hostname, resolver, logger)
					success = append(success, s...)
					failed = append(failed, f...)
				}
			}
		}

		unwrappedSuccess, unwrappedFail := unwrapLookups(success, failed)

		if len(success) > 0 {
			lookups = append(success, failed...)
		} else {
			_, _ = fmt.Fprintf(logger, "Unable to detect any LANCache instances %s\n\n", resolverMsg)
			if debug {
				_, _ = fmt.Fprintf(logfile, "Successful lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s\n", len(success), unwrappedSuccess, len(failed), unwrappedFail)
			}
			continue
		}

		deltas = isLookupInSliceEqual(lookups)
		logOutput(host, resolverMsg, unwrappedSuccess, unwrappedFail, hostnames, iterations, lookups, success, failed, deltas, logger, logfile, debug)
	}
}

func processHostnames(hostname, resolver string, logger io.Writer) (success, failed []Lookup) {
	ips, transport, err := resolveIP(hostname, resolver+portDNS)
	if len(ips) == 0 {
		failed = append(failed, Lookup{
			Resolver: resolver,
			Hostname: hostname,
			Time:     time.Now().Format(time.RFC822),
		})
		return success, failed
	}
	if err != nil {
		_, _ = fmt.Fprintf(logger, "Could not get IPs: %v\n", err)
		failed = append(failed, Lookup{
			Resolver: resolver,
			Hostname: hostname,
			Time:     time.Now().Format(time.RFC822),
		})
		return success, failed
	}

	success, failed = lookupHeartbeat(hostname, resolver+portDNS, ips, transport)

	return success, failed
}

func lookupHeartbeat(hostname, resolver string, ips []string, transport *http.Transport) (success, failed []Lookup) {
	client := &http.Client{
		Timeout:   1 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(httpPrefix + hostname + heartbeatSuffix)
	if err != nil {
		failed = append(failed, Lookup{
			Resolver: resolver,
			Hostname: hostname,
			Address:  ips,
			Time:     time.Now().Format(time.RFC822),
		})
		return success, failed
	}

	if resp.Header.Get(lancacheHeader) != "" {
		success = append(success, Lookup{
			Resolver:    resolver,
			Hostname:    hostname,
			Address:     ips,
			ContainerID: resp.Header.Get(lancacheHeader),
			Time:        time.Now().Format(time.RFC822),
		})
	} else {
		failed = append(failed, Lookup{
			Resolver: resolver,
			Hostname: hostname,
			Address:  ips,
			Time:     time.Now().Format(time.RFC822),
		})
	}

	return success, failed
}

func parseCDN(name, file string, logger io.Writer) (hostnames []string) {
	cdnHosts, err := urlToLines(cacheRepo+file, logger)
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: failed to parse cdn file %w", err))
	}

	_, _ = fmt.Fprintf(logger, "-----------------------------------------------------------------\n"+
		"Looking up CDN: %s diagnostics addresses...\n"+
		"-----------------------------------------------------------------\n", name)
	for _, host := range cdnHosts {
		host = strings.TrimSpace(host)
		if strings.HasPrefix(host, "#") {
			continue
		}
		if strings.HasPrefix(host, "*.") {
			host = strings.Replace(host, wildcardPrefix, testPrefix, 1)
		}
		hostnames = append(hostnames, host)
	}

	return hostnames
}

func resolveIP(hostname, resolver string) ([]string, *http.Transport, error) {
	var (
		dialer net.Dialer
		ips    []string
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if resolver == systemResolver[0]+portDNS {
		ips, err := net.LookupHost(hostname)
		return ips, &http.Transport{}, err
	}

	// Fixed in Go 1.19: https://github.com/golang/go/issues/33097
	//r := &net.Resolver{
	//	PreferGo: true,
	//	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
	//		dialer = net.Dialer{
	//			Timeout: 1 * time.Second,
	//		}
	//		return dialer.DialContext(ctx, network, resolver+portDNS)
	//	},
	//}
	//
	//ips, err = r.LookupHost(context.Background(), hostname)

	client := new(dns.Client)
	parsed := net.ParseIP(hostname)
	if parsed != nil {
		ips = append(ips, hostname)
		return ips, &http.Transport{}, nil
	}

	messageAAAA := new(dns.Msg)
	messageAAAA.SetQuestion(dns.Fqdn(hostname), dns.TypeAAAA)
	inAAAA, _, err := client.ExchangeContext(ctx, messageAAAA, resolver)

	if err != nil {
		return nil, &http.Transport{}, err
	}

	for _, record := range inAAAA.Answer {
		if t, ok := record.(*dns.AAAA); ok {
			ips = append(ips, t.AAAA.String())
		}
	}

	messageA := new(dns.Msg)
	messageA.SetQuestion(dns.Fqdn(hostname), dns.TypeA)
	inA, _, err := client.ExchangeContext(ctx, messageA, resolver)

	if err != nil {
		return nil, &http.Transport{}, err
	}

	for _, record := range inA.Answer {
		if t, ok := record.(*dns.A); ok {
			ips = append(ips, t.A.String())
		}
	}

	transport := http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		addr = ips[0] + portHTTP
		return dialer.DialContext(ctx, network, addr)
	}}

	return ips, &transport, nil
}

func logOutput(host, resolverMsg, unwrappedSuccess, unwrappedFail string, hostnames []string, iterations int, lookups, success, failed, deltas []Lookup, logger io.Writer, logfile *os.File, debug bool) {
	if len(deltas) > 0 {
		first := lookups[0]

		if host != "" {
			_, _ = fmt.Fprintf(logger, "Unsuccessfully ran %d diagnostics iteration(s) %s\n"+
				"\nSuccessful lookups: %d\n"+
				"%s"+
				"\nFailed lookups: %d\n"+
				"%s"+
				"\nDidn't match:\n"+
				"%+v\n\n", iterations, resolverMsg, len(success), unwrappedSuccess, len(failed), unwrappedFail, first)
		} else {
			_, _ = fmt.Fprintf(logger, "Unsuccessfully ran %d diagnostics iteration(s) on %d host(s) %s\n"+
				"\nSuccessful lookups: %d\n"+
				"%s"+
				"\nFailed lookups: %d\n"+
				"%s"+
				"\nDidn't match:\n"+
				"%+v\n\n", iterations, len(hostnames), resolverMsg, len(success), unwrappedSuccess, len(failed), unwrappedFail, first)
		}
	} else {
		if host != "" {
			_, _ = fmt.Fprintf(logger, "Successfully ran %d diagnostics iteration(s) %s\n\n", iterations, resolverMsg)
		} else {
			_, _ = fmt.Fprintf(logger, "Successfully ran %d diagnostics iteration(s) on %d host(s) %s\n\n", iterations, len(hostnames), resolverMsg)
		}
		if debug {
			_, _ = fmt.Fprintf(logfile, "Successful lookups: %d\n"+
				"%s"+
				"\nFailed lookups: %d\n"+
				"%s\n", len(success), unwrappedSuccess, len(failed), unwrappedFail)
		}
	}
}
