package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	ciliumApiPolicy "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumClientSet "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	"github.com/cilium/cilium/pkg/policy/api"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func writeToFile(filename string, data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		log.Fatalf("Failed to marshal ciliumPolicy to YAML: %v", err)
		return err
	}

	err = os.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
		return err
	}

	fmt.Println("data written")
	return nil
}

func createNetPol(policyNamespace string, ciliumPolicy ciliumApiPolicy.CiliumNetworkPolicy) error {
	// uses the current context in the kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	ciliumClient, err := ciliumClientSet.NewForConfig(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = ciliumClient.CiliumV2().CiliumNetworkPolicies(policyNamespace).Create(ctx, &ciliumPolicy, metaV1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var egressRule []api.EgressRule
	var err error

	for _, link := range ipRangeURL {
		egressRule, err = generateGcpEgressRule(link)
		if err != nil {
			fmt.Printf("ERROR: Could not get data from %s\n", link)
			return
		}
	}
	fmt.Println(egressRule)
}
