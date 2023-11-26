package main

import (
	"context"
	"encoding/json"
	"fmt"
	ciliumApiPolicy "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumClientSet "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	slimMetaV1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	ciliumApiRules "github.com/cilium/cilium/pkg/policy/api"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"log"
	"os"
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
	egressRule, err := generateGithubEgressRule()
	if err != nil {
		return
	}

	ruleNamespace := "arc-runners"
	ruleName := "ingress-egress-github"

	spec := ciliumApiRules.Rule{
		EndpointSelector: ciliumApiRules.EndpointSelector{
			LabelSelector: &slimMetaV1.LabelSelector{},
		},
		Egress: egressRule,
	}

	ciliumPolicy := ciliumApiPolicy.CiliumNetworkPolicy{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "CiliumNetworkPolicy",
			APIVersion: "cilium.io/v2",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      ruleName,
			Namespace: ruleNamespace,
		},
		Spec:   &spec,
		Specs:  nil,
		Status: ciliumApiPolicy.CiliumNetworkPolicyStatus{},
	}

	err = writeToFile("ciliumPolicy.yaml", ciliumPolicy)
	if err != nil {
		panic(err)
	}

	err = createNetPol(ruleNamespace, ciliumPolicy)
	if err != nil {
		panic(err)
	}
}
