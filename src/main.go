package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

// Variables set during build
var (
	ProjectName  string
	BuildVersion string
	BuildDate    string
)

var statusMap = []string{
	"OK",
	"WARN",
	"CRIT",
	"UNKNOWN",
}

type esStatus string

const (
	esStatusGreen  esStatus = "green"
	esStatusYellow          = "yellow"
	esStatusRed             = "red"
)

type esHealthResponse struct {
	ClusterName string   `json:"cluster_name"`
	Status      esStatus `json:"status"`
}

var (
	flagVersion = flag.Bool("v", false, "Print the version info and exit")
	flagService = flag.String("service", "", "Service name (defaults to Elasticsearch_<host>)")
	flagHost    = flag.String("host", "", "Host")
	flagPort    = flag.Int("port", 9200, "Port")
	flagDNS     = flag.String("dns", "", "Use alternate dns server")
)

func resolveDNS(host string) (string, error) {
	c := dns.Client{}
	m := dns.Msg{}

	m.SetQuestion(host+".", dns.TypeA)

	r, _, err := c.Exchange(&m, *flagDNS)
	if err != nil {
		return "", fmt.Errorf("Can't resolve '%s' on %s: %s", host, *flagDNS, err)
	}

	if len(r.Answer) == 0 {
		return "", fmt.Errorf("Can't resolve '%s' on %s: No results", host, *flagDNS)
	}

	aRecord := r.Answer[0].(*dns.A)

	return aRecord.A.String(), nil
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Printf("%s %s (Build %s)\n", ProjectName, BuildVersion, BuildDate)
		fmt.Printf("\n")
		fmt.Printf("https://github.com/indece-official/sshmon-check-elasticsearch\n")
		fmt.Printf("\n")
		fmt.Printf("Copyright 2020 by indece UG (haftungsbeschränkt)\n")

		os.Exit(0)

		return
	}

	serviceName := *flagService
	if serviceName == "" {
		serviceName = fmt.Sprintf("Elasticsearch_%s", *flagHost)
	}

	url := fmt.Sprintf("http://%s:%d/_cluster/health", *flagHost, *flagPort)

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				var err error

				if *flagDNS != "" {
					addrParts := strings.Split(addr, ":")
					if len(addrParts) != 2 {
						return nil, fmt.Errorf("Error parsing address '%s': must have format <host>:<port>", addr)
					}

					addrParts[0], err = resolveDNS(addrParts[0])
					if err != nil {
						return nil, err
					}

					addr = strings.Join(addrParts, ":")
				}

				return net.Dial(network, addr)
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf(
			"2 %s - %s - Error getting health from '%s': %s\n",
			serviceName,
			statusMap[2],
			*flagHost,
			err,
		)

		os.Exit(1)

		return
	}
	defer resp.Body.Close()

	esResp := esHealthResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(
			"2 %s - %s - Error reading response body from health request to '%s': %s\n",
			serviceName,
			statusMap[2],
			*flagHost,
			err,
		)

		os.Exit(1)

		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf(
			"2 %s - %s - Error getting health from '%s' - %s\n",
			serviceName,
			statusMap[2],
			*flagHost,
			resp.Status,
		)

		os.Exit(1)

		return
	}

	err = json.Unmarshal(body, &esResp)
	if err != nil {
		fmt.Printf(
			"2 %s - %s - Error parsing response body from health request to '%s': %s\n",
			serviceName,
			statusMap[2],
			*flagHost,
			err,
		)

		os.Exit(1)

		return
	}

	status := 0
	switch esResp.Status {
	case esStatusGreen:
		status = 0
	case esStatusYellow:
		status = 1
	case esStatusRed:
		status = 2
	}

	fmt.Printf(
		"%d %s - %s - Elasticsearch cluster '%s' on %s has status '%s'\n",
		status,
		serviceName,
		statusMap[status],
		esResp.ClusterName,
		*flagHost,
		esResp.Status,
	)

	os.Exit(0)
}
