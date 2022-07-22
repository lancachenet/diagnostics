package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

func getInterfaceAddresses(logger io.Writer) {
	interfaces, err := net.Interfaces()
	if err != nil {
		_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
		return
	}

	_, _ = fmt.Fprintf(logger, "\n")

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
				_, _ = fmt.Fprintf(logfile, "Successfull lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s\n", len(success), unwrappedSuccess, len(failed), unwrappedFail)
			}
			continue
		}

		deltas = isLookupInSliceEqual(lookups)
		if len(deltas) > 0 {
			first := lookups[0]

			if host != "" {
				_, _ = fmt.Fprintf(logger, "Unsuccessfully ran %d diagnostics iteration(s) %s\n"+
					"\nSuccessfull lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s"+
					"\nDidn't match:\n"+
					"%+v\n\n", iterations, resolverMsg, len(success), unwrappedSuccess, len(failed), unwrappedFail, first)
			} else {
				_, _ = fmt.Fprintf(logger, "Unsuccessfully ran %d diagnostics iteration(s) on %d host(s) %s\n"+
					"\nSuccessfull lookups: %d\n"+
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
				_, _ = fmt.Fprintf(logfile, "Successfull lookups: %d\n"+
					"%s"+
					"\nFailed lookups: %d\n"+
					"%s\n", len(success), unwrappedSuccess, len(failed), unwrappedFail)
			}
		}
	}
}

func processHostnames(hostname, resolver string, logger io.Writer) (success, failed []Lookup) {
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
		_, _ = fmt.Fprintf(logger, "Could not get IPs: %v\n", err)
		failed = append(failed, Lookup{
			Resolver: resolver,
			Hostname: hostname,
			Time:     time.Now().String(),
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

func isLookupInSliceEqual(a []Lookup) []Lookup {
	var l []Lookup

	for _, v := range a {
		original := v
		v.Hostname = a[0].Hostname
		v.Time = a[0].Time
		if !reflect.DeepEqual(v, a[0]) {
			l = append(l, original)
		}
	}

	return l
}

func unwrapLookups(s, f []Lookup) (success, fail string) {
	for _, s := range s {
		success += fmt.Sprintf("+%+v\n", s)
	}
	for _, f := range f {
		fail += fmt.Sprintf("-%+v\n", f)
	}

	return success, fail
}

func urlToLines(url string, logger io.Writer) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_, _ = fmt.Fprint(logger, fmt.Errorf("error: %w", err))
		}
	}(resp.Body)

	return linesFromReader(resp.Body)
}

func linesFromReader(r io.Reader) ([]string, error) {
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
