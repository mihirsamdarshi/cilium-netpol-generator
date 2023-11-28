package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cilium/cilium/pkg/policy/api"
)

type IPRange struct {
	Prefixes []string `json:"prefixes"`
}

var (
	ipRangeURL = []string{
		"https://www.gstatic.com/ipranges/goog.json",
		"https://www.gstatic.com/ipranges/cloud.json",
	}
)

func generateGcpEgressRule(link string) ([]api.EgressRule, error) {
	var ipRange IPRange
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("ERROR: Invalid HTTP response from %s\n", link)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("ERROR: Could not close HTTP response body from %s\n", link)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("ERROR: Invalid HTTP response code from %s\n", link)
		return nil, err
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	// print body bytes to string
	bodyString := string(bodyBytes)
	fmt.Printf("Body: %s\n", bodyString)
	err = json.Unmarshal(bodyBytes, &ipRange)
	if err != nil {
		fmt.Printf("ERROR: Could not unmarshal JSON from %s\n", link)
		return nil, err
	}
	// Convert the list of IP strings to a list of CIDRs (CIDR is a type alias for string)
	cidrList := make([]api.CIDR, len(ipRange.Prefixes))
	for i, ip := range ipRange.Prefixes {
		cidrList[i] = api.CIDR(ip)
	}

	portList := make([]api.PortProtocol, 3)
	for i, port := range []int{22, 80, 443} {
		portList[i] = api.PortProtocol{
			Port:     fmt.Sprintf("%d", port),
			Protocol: "TCP",
		}
	}
	portRuleList := []api.PortRule{
		{
			Ports: portList,
		},
	}

	egressRule := []api.EgressRule{
		{
			EgressCommonRule: api.EgressCommonRule{
				ToCIDR: cidrList,
			},
			ToPorts: portRuleList,
		},
	}

	return egressRule, nil
}
