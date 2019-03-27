package common

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	"github.com/sirupsen/logrus"
)

// RestorePlugin is a restore item action plugin for Heptio Ark.
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything.
func (p *RestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{}, nil
}

// Execute sets a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	p.Log.Info("Hello from common restore plugin!!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	client, err := clients.NewDiscoveryClient()
	if err != nil {
		return nil, nil, err
	}
	version, err := client.ServerVersion()
	if err != nil {
		return nil, nil, err
	}
	if strings.HasSuffix(version.Minor, "+") {
		version.Minor = strings.TrimSuffix(version.Minor, "+")
	}

	annotations[RestoreServerVersion] = fmt.Sprintf("%v.%v", version.Major, version.Minor)
	registryHostname, err := getRegistryInfo(version.Major, version.Minor)
	if err != nil {
		return nil, nil, err
	}
	annotations[RestoreRegistryHostname] = registryHostname
	metadata.SetAnnotations(annotations)

	return item, nil, nil
}
