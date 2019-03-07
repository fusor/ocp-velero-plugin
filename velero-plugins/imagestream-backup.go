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

package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"strings"

	imagev1API "github.com/openshift/api/image/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/backup"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type ImageStreamBackupPlugin struct {
	log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *ImageStreamBackupPlugin) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{
		IncludedResources: []string{"imagestreams"},
	}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *ImageStreamBackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.log.Info("Hello from Imagestream backup plugin!!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["openshift.io/imagestream-plugin"] = "1"

	im := imagev1API.ImageStream{}
	obj := item.UnstructuredContent()
	mapstructure.Decode(obj, &im)
	p.log.Info(fmt.Sprintf("image: %#v", im.Status))
	dockerRepo := im.Status.DockerImageRepository

	// Get associated image and export to scratch location
	client, err := p.imageClient()
	if err != nil {
		return nil, nil, err
	}
	images, err := client.Images().List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	for _, image := range images.Items {
		repo := strings.Split(image.DockerImageReference, "@")[0]
		if repo == dockerRepo {
			annotations["openshift.io/dockerImageRepo"] = repo
		}
	}
	metadata.SetAnnotations(annotations)

	return item, nil, nil
}

func (p *ImageStreamBackupPlugin) imageClient() (*imagev1.ImageV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := imagev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
