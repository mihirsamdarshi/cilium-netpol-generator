package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

type IPRange struct {
	Prefixes []IPPrefix `json:"prefixes"`
}

type IPPrefix struct {
	IPv4Prefix string `json:"ipv4Prefix"`
	IPv6Prefix string `json:"ipv6Prefix"`
}

type CIDR struct {
	IPv4CIDR *net.IPNet
	IPv6CIDR *net.IPNet
}

var (
	ipRangeURL = map[string]string{
		"goog":  "https://www.gstatic.com/ipranges/goog.json",
		"cloud": "https://www.gstatic.com/ipranges/cloud.json",
	}
)

func getData(link string) (CIDR, string) {
	var ipRange IPRange
	var creationTime string
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("ERROR: Invalid HTTP response from %s\n", link)
		return CIDR{}, ""
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("ERROR: Could not close HTTP response body from %s\n", link)
			return
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(bodyBytes, &ipRange)
		if err != nil {
			fmt.Printf("ERROR: Could not unmarshal JSON from %s\n", link)
			return CIDR{}, ""
		}
		for _, prefix := range ipRange.Prefixes {
			_, ipnet, _ := net.ParseCIDR(prefix.IPv4Prefix)
			_, ipnet6, _ := net.ParseCIDR(prefix.IPv6Prefix)

			return CIDR{
				IPv4CIDR: ipnet,
				IPv6CIDR: ipnet6,
			}, creationTime
		}
	}

	return CIDR{}, ""
}

func main() {
	var cidrs = make(map[string]CIDR)

	for group, link := range ipRangeURL {
		cidrs[group], _ = getData(link)
	}
	if len(cidrs) != 2 {
		fmt.Println("ERROR: Could process data from Google")
	}
	fmt.Println("IP ranges for Google APIs and services default domains:")
	for ip, _ := range cidrs {
		fmt.Println(ip)
	}
}
