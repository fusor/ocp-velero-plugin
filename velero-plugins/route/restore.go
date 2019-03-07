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
	"fmt"
	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	networkv1 "github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

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

	client, err := p.networkClient()
	if err != nil {
		return nil, nil, err
	}
	subnets, err := client.HostSubnets().List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	clusternetworks, err := client.ClusterNetworks().List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	p.Log.Info(fmt.Sprintf("subnet: %#v", subnets))
	p.Log.Info(fmt.Sprintf("clusterNetworks: %#v", clusternetworks))

	return item, nil, nil
}

func (p *RestorePlugin) networkClient() (*networkv1.NetworkV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := networkv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
