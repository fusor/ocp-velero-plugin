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

package build

import (
	"encoding/json"
	"fmt"
	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	buildv1API "github.com/openshift/api/build/v1"
	"github.com/sirupsen/logrus"
	corev1API "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strings"
)

// MyRestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything
func (p *RestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"builds"},
	}, nil
}

func (p *RestorePlugin) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	p.Log.Info("Hello from Build RestorePlugin!")

	build := buildv1API.Build{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &build)
	if build.Spec.Strategy.Type == buildv1API.SourceBuildStrategyType {
		secret := p.findBuilderDockercfgSecret(build.Namespace)
		if secret == "" {
			// TODO: Come back to this. This is ugly, should really return some type
			// of error but I don't know what that is exactly
			return item, nil, nil
		}
		p.Log.Info(fmt.Sprintf("Found new dockercfg secret: %v", secret))

		newPushSecret := corev1API.LocalObjectReference{Name: secret}
		build.Spec.Output.PushSecret = &newPushSecret
		build.Spec.Strategy.SourceStrategy.PullSecret = &newPushSecret
	}
	var out map[string]interface{}
	objrec, _ := json.Marshal(build)
	json.Unmarshal(objrec, &out)
	item.SetUnstructuredContent(out)

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

func (p *RestorePlugin) findBuilderDockercfgSecret(namespace string) string {
	client, err := p.coreClient()
	if err != nil {
		return ""
	}
	secretList, err := client.Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return ""
	}
	for _, secret := range secretList.Items {
		if strings.HasPrefix(secret.Name, "builder-dockercfg") {
			return secret.Name
		}
	}
	return ""
}
