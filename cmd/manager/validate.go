package main

import (
	"fmt"

	version "github.com/hashicorp/go-version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

const (
	// MinKubeApiVersio is the lowest Kubernetes API version splunk operator supports.
	// version string should be the same as the git release tag
	MinKubeApiVersion = "v1.14.8"

	// MaxKubeApiVersion is the highest Kubernetes API version splunk operator supports.
	MaxKubeApiVersion = "v1.16.2"
)

func CheckSupportedK8sVersion(c *rest.Config) error {
	versionClient := discovery.NewDiscoveryClientForConfigOrDie(c)
	serverApiVersion, err := versionClient.ServerVersion()
	if err != nil {
		return err
	}

	serverVersion, err := version.NewVersion(serverApiVersion.String())
	if err != nil {
		return err
	}

	minVersion, err := version.NewVersion(MinKubeApiVersion)
	if err != nil {
		return err
	}
	maxVersion, err := version.NewVersion(MaxKubeApiVersion)
	if err != nil {
		return err
	}
	if serverVersion.LessThan(minVersion) {
		return fmt.Errorf("minimum supported Kubernetes version is %s, but the server version is %s.", minVersion, serverVersion.String())
	}

	if serverVersion.GreaterThan(maxVersion) {
		return fmt.Errorf("The maximum supported Kubernetes version is %s, but the server version is %s.", maxVersion, serverVersion.String())
	}

	return nil
}
