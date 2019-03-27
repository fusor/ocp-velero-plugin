package buildconfig

import (
	"github.com/sirupsen/logrus"

	//buildv1API "github.com/openshift/api/build/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
