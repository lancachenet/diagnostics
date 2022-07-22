package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func getInterfaceAddresses() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("error: %w", err))
		return
	}

	fmt.Printf("\n")

	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("error: %w", err))
			continue
		}

		if strings.Contains(i.Flags.String(), up) {
			if !strings.Contains(strings.ToLower(i.Name), loopback) && !strings.Contains(i.Flags.String(), loopback) {
				fmt.Printf("Interface: %s\n", i.Name)

				for _, a := range addresses {
					switch v := a.(type) {
					case *net.IPAddr:
						fmt.Printf("IP Address: %s\n", v)

					case *net.IPNet:
						fmt.Printf("IP Address: %s\n", v)
					}
				}
				fmt.Printf("\n")
			}
		}
	}
}

func lookupHostnames(host string, hostnames []string, iterations int, servers []string) {
	var (
		lookups []Lookup
		success []Lookup
		failed  []Lookup
		deltas  []Lookup
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
				s, f := processHostnames(host, "")
				success = append(success, s...)
				failed = append(failed, f...)
			} else {
				for _, hostname := range hostnames {
					s, f := processHostnames(hostname, resolver)
					success = append(success, s...)
					failed = append(failed, f...)
				}
			}
		}

		if len(success) > 0 {
			lookups = append(success, failed...)
		} else {
			fmt.Printf("Unable to detect any LANCache instances\n\n")
			return
		}

		if host != "" {
			deltas = isLookupInSliceEqual(lookups, false)
		} else {
			deltas = isLookupInSliceEqual(lookups, true)
		}

		if len(deltas) > 0 {
			unwrappedSuccess := ""
			unwrappedFail := ""
			for _, s := range success {
				unwrappedSuccess += fmt.Sprintf("+%+v\n", s)
			}
			for _, f := range failed {
				unwrappedFail += fmt.Sprintf("-%+v\n", f)
			}

			first := lookups[0]

			if host != "" {
				fmt.Printf("Unsuccessfully ran %d diagnostics iterations %s\n"+
					"\nSuccessfull lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s"+
					"\nDidn't match:\n"+
					"%+v\n\n", iterations, resolverMsg, len(success), unwrappedSuccess, len(failed), unwrappedFail, first)
			} else {
				fmt.Printf("Unsuccessfully ran %d diagnostics iterations on %d hosts %s\n"+
					"\nSuccessfull lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s"+
					"\nDidn't match:\n"+
					"%+v\n\n", iterations, len(hostnames), resolverMsg, len(success), unwrappedSuccess, len(failed), unwrappedFail, first)
			}
		} else {
			if host != "" {
				fmt.Printf("Successfully ran %d diagnostics iterations %s\n\n", iterations, resolverMsg)
			} else {
				fmt.Printf("Successfully ran %d diagnostics iterations on %d hosts %s\n\n", iterations, len(hostnames), resolverMsg)
			}
		}
	}
}

func processHostnames(hostname, resolver string) (success, failed []Lookup) {
	var (
		ips []net.IP
		err error
	)

	if resolver != "system" {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 1 * time.Second,
				}
				return d.DialContext(ctx, network, resolver)
			},
		}

		ips, err = r.LookupIP(context.Background(), "ip", hostname)
	} else {
		ips, err = net.LookupIP(hostname)
	}

	if err != nil {
		fmt.Printf("Could not get IPs: %v\n", err)
		failed = append(failed, Lookup{
			Hostname: hostname,
		})
		return success, failed
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(httpPrefix + hostname + heartbeatSuffix)
	if err != nil {
		failed = append(failed, Lookup{
			Hostname: hostname,
			Address:  ips,
		})
		return success, failed
	}

	if resp.Header.Get(lancacheHeader) != "" {
		success = append(success, Lookup{
			Hostname:    hostname,
			Address:     ips,
			ContainerID: resp.Header.Get(lancacheHeader),
		})
	} else {
		failed = append(failed, Lookup{
			Hostname: hostname,
			Address:  ips,
		})
	}

	return success, failed
}

func parseCDN(name, file string) (hostnames []string) {
	cdnHosts, err := URLToLines(cacheRepo + file)
	if err != nil {
		fmt.Print(fmt.Errorf("error: failed to parse cdn file %w", err))
	}

	fmt.Printf("Looking up CDN: %s diagnostics addresses...\n", name)
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

func isLookupInSliceEqual(a []Lookup, multihost bool) []Lookup {
	var l []Lookup

	for _, v := range a {
		original := v
		if multihost {
			v.Hostname = a[0].Hostname
		}
		if !reflect.DeepEqual(v, a[0]) {
			if multihost {
				l = append(l, original)
			} else {
				l = append(l, v)
			}
		}
	}

	return l
}

func isStringInSlice(needle string, haystack []string) (inSlice bool) {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}

	return false
}

func URLToLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(fmt.Errorf("error: %w", err))
		}
	}(resp.Body)

	return LinesFromReader(resp.Body)
}

func LinesFromReader(r io.Reader) ([]string, error) {
	var lines []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
