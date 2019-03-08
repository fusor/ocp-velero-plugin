/*
Copyright 2018 the Heptio Ark contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package route

import (
	"encoding/json"
	"fmt"
	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	"github.com/mitchellh/mapstructure"
	routev1API "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strings"
)

type routingConfig struct {
	Subdomain string `json:"subdomain"`
}

type imagePolicyConfig struct {
	InternalRegistryHostname string `json:"internalRegistryHostname"`
}

type ApiServerConfig struct {
	ImagePolicyConfig imagePolicyConfig `json:"imagePolicyConfig"`
	RoutingConfig     routingConfig     `json:"routingConfig"`
}

// MyRestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything
func (p *RestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"routes"},
	}, nil
}

func (p *RestorePlugin) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	p.Log.Info("Hello from Route RestorePlugin!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["openshift.io/route-restore-plugin"] = "1"

	metadata.SetAnnotations(annotations)

	client, err := p.coreClient()
	if err != nil {
		p.Log.Info(fmt.Sprintf("ERROR: %v", err))
		return nil, nil, err
	}
	config, err := client.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
	if err != nil {
		p.Log.Info(fmt.Sprintf("ERROR: %v", err))
		return nil, nil, err
	}
	serverConfig := ApiServerConfig{}
	err = json.Unmarshal([]byte(config.Data["config.yaml"]), &serverConfig)
	if err != nil {
		p.Log.Info(fmt.Sprintf("ERROR: %v", err))
		return nil, nil, err
	}

	subdomain := serverConfig.RoutingConfig.Subdomain

	route := routev1API.Route{}
	obj := item.UnstructuredContent()
	mapstructure.Decode(obj, &route)

	host := route.Spec.Host
	name := strings.Split(host, ".")[0]
	newHost := name + "." + subdomain
	route.Spec.Host = newHost
	var out map[string]interface{}
	objrec, _ := json.Marshal(route)
	json.Unmarshal(objrec, &out)

	item.SetUnstructuredContent(out)
	p.Log.Info(fmt.Sprintf("New route: %v", newHost))
	p.Log.Info(fmt.Sprintf("item: %#v", item))

	return item, nil, nil
}

func (p *RestorePlugin) coreClient() (*corev1.CoreV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := corev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
