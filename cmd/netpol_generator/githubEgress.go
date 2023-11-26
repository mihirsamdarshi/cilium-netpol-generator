package main

import (
	"encoding/json"
	"fmt"
	"github.com/cilium/cilium/pkg/policy/api"
	"io"
	"log"
	"net/http"
	"strings"
)

const GithubMetaUrl = "https://api.github.com/meta"

type SshKeyFingerprints struct {
	SHA256ECDSA   string `json:"SHA256_ECDSA"`
	SHA256ED25519 string `json:"SHA256_ED25519"`
	SHA256RSA     string `json:"SHA256_RSA"`
}

type Domains struct {
	Website    []string `json:"website"`
	Codespaces []string `json:"codespaces"`
	Copilot    []string `json:"copilot"`
	Packages   []string `json:"packages"`
}

type GitHubMetaApiResponse struct {
	VerifiablePasswordAuthentication bool `json:"verifiable_password_authentication"`
	SshKeyFingerprints               `json:"ssh_key_fingerprints"`
	SshKeys                          []string `json:"ssh_keys"`
	Hooks                            []string `json:"hooks"`
	Web                              []string `json:"web"`
	Api                              []string `json:"api"`
	Git                              []string `json:"git"`
	GithubEnterpriseImporter         []string `json:"github_enterprise_importer"`
	Packages                         []string `json:"packages"`
	Pages                            []string `json:"pages"`
	Importer                         []string `json:"importer"`
	Actions                          []string `json:"actions"`
	Dependabot                       []string `json:"dependabot"`
	Domains                          `json:"domains"`
}

// getGithubMeta fetches the GitHub meta API and returns the response
func getGithubMeta() (GitHubMetaApiResponse, error) {
	res, err := http.Get(GithubMetaUrl)
	if err != nil {
		return GitHubMetaApiResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	var metaApi GitHubMetaApiResponse
	err = json.NewDecoder(res.Body).Decode(&metaApi)
	if err != nil {
		return GitHubMetaApiResponse{}, err
	}

	log.Println("Received GitHub Meta API response")
	return metaApi, nil
}

func generateGithubEgressRule() ([]api.EgressRule, error) {
	// Get the GitHub meta API meta_api
	metaRes, err := getGithubMeta()
	if err != nil {
		log.Fatalf("Failed to get Github Meta: %v", err)
		return nil, err
	}

	// unpack the lists of IPs into a single array
	mergedIpAddrList := NewSet().AddLists(
		metaRes.Hooks,
		metaRes.Web,
		metaRes.Api,
		metaRes.Git,
		metaRes.GithubEnterpriseImporter,
		metaRes.Packages,
		metaRes.Pages,
		metaRes.Importer,
		metaRes.Actions,
		metaRes.Dependabot).ToList()

	mergedDomainList := NewSet().AddLists(
		metaRes.Domains.Website,
		metaRes.Domains.Codespaces,
		metaRes.Domains.Copilot,
		metaRes.Domains.Packages).ToList()

	// Convert the list of IP strings to a list of CIDRs (CIDR is a type alias for string)
	addrList := make([]api.CIDR, len(mergedIpAddrList))
	for i, ip := range mergedIpAddrList {
		addrList[i] = api.CIDR(ip)
	}

	cidrList := make([]api.FQDNSelector, len(mergedDomainList))
	for i, domain := range mergedDomainList {
		if strings.HasPrefix(domain, "*.") {
			cidrList[i] = api.FQDNSelector{
				MatchPattern: domain,
			}
		} else {
			cidrList[i] = api.FQDNSelector{
				MatchName: domain,
			}
		}

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
				ToCIDR: addrList,
			},
			ToPorts: portRuleList,
		},
		{
			ToFQDNs: cidrList,
			ToPorts: portRuleList,
		},
	}

	return egressRule, nil
}
