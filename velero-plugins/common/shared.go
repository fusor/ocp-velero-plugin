package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRegistryInfo(major, minor string) (string, error) {
	if major != "1" {
		return "", fmt.Errorf("server version %v.%v not supported. Must be 1.x", major, minor)
	}
	intVersion, err := strconv.Atoi(minor)
	if err != nil {
		return "", fmt.Errorf("server minor version %v invalid value: %v", minor, err)
	}

	cClient, err := clients.NewCoreClient()
	if err != nil {
		return "", err
	}
	if intVersion < 7 {
		return "", fmt.Errorf("Kubernetes version 1.%v not supported. Must be 1.7 or greater", minor)
	} else if intVersion <= 11 {
		registrySvc, err := cClient.Services("default").Get("docker-registry", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		internalRegistry := registrySvc.Spec.ClusterIP + ":" + strconv.Itoa(int(registrySvc.Spec.Ports[0].Port))
		return internalRegistry, nil
	} else {
		config, err := cClient.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		serverConfig := APIServerConfig{}
		err = json.Unmarshal([]byte(config.Data["config.yaml"]), &serverConfig)
		if err != nil {
			return "", err
		}
		internalRegistry := serverConfig.ImagePolicyConfig.InternalRegistryHostname
		if len(internalRegistry) == 0 {
			return "", errors.New("InternalRegistryHostname not found")
		}
		return internalRegistry, nil
	}
}
