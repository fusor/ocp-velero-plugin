/*
Copyright 2017 the Heptio Ark contributors.

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

package imagestream

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"strings"

	buildv1API "github.com/openshift/api/build/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/backup"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{
		IncludedResources: []string{"buildconfigs"},
	}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.Log.Info("Hello from Build Config backup plugin!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["openshift.io/buildconfig-plugin"] = "1"

	/*client, err := p.buildClient()
	if err != nil {
		return nil, nil, err
	}*/
	metadata.SetAnnotations(annotations)

	return item, nil, nil
}

func (p *BackupPlugin) buildClient() (*buildv1.BuildV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := buildv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
