package common

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/backup"
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

	client, err := p.discoveryClient()
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

	annotations["openshift.io/backup-server-version"] = fmt.Sprintf("%v.%v", version.Major, version.Minor)

	metadata.SetAnnotations(annotations)

	return item, nil, nil
}

func (p *BackupPlugin) discoveryClient() (*discovery.DiscoveryClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
