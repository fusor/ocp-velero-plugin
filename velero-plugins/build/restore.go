package build

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	v1 "github.com/heptio/velero/pkg/apis/ark/v1"
	"github.com/heptio/velero/pkg/restore"
	buildv1API "github.com/openshift/api/build/v1"
	"github.com/sirupsen/logrus"
	corev1API "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything
func (p *RestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"builds"},
	}, nil
}

// Execute action for the restore plugin for the build resource
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

func (p *RestorePlugin) findBuilderDockercfgSecret(namespace string) string {
	client, err := clients.NewCoreClient()
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
