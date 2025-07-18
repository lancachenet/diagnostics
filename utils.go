package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

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

func (l Lookup) String() string {
	return fmt.Sprintf("%s %s %v", l.Hostname, l.Resolver, l.Address)
}

func unwrapLookups(s, f []Lookup) (success, fail string) {
	for _, lookup := range s {
		success += "+" + lookup.String() + "\n"
	}
	for _, lookup := range f {
		fail += "-" + lookup.String() + "\n"
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
