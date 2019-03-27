package common

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/backup"
	"github.com/sirupsen/logrus"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.Log.Info("Hello from common backup plugin!!")

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

	annotations[BackupServerVersion] = fmt.Sprintf("%v.%v", version.Major, version.Minor)
	registryHostname, err := getRegistryInfo(version.Major, version.Minor)
	if err != nil {
		return nil, nil, err
	}
	annotations[BackupRegistryHostname] = registryHostname
	metadata.SetAnnotations(annotations)

	return item, nil, nil
}
