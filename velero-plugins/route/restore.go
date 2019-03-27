package route

import (
	"encoding/json"
	"strings"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	v1 "github.com/heptio/velero/pkg/apis/ark/v1"
	"github.com/heptio/velero/pkg/restore"
	routev1API "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
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
		IncludedResources: []string{"routes"},
	}, nil
}

// Execute fixes the route path on restore to use the target cluster's domain name
func (p *RestorePlugin) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	p.Log.Info("Hello from Route RestorePlugin!")
	route := routev1API.Route{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &route)

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

	client, err := clients.NewCoreClient()
	if err != nil {
		return nil, nil, err
	}
	config, err := client.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	serverConfig := common.APIServerConfig{}
	err = json.Unmarshal([]byte(config.Data["config.yaml"]), &serverConfig)
	if err != nil {
		return nil, nil, err
	}

	subdomain := serverConfig.RoutingConfig.Subdomain

	host := route.Spec.Host
	name := strings.Split(host, ".")[0]
	newHost := name + "." + subdomain
	route.Spec.Host = newHost

	var out map[string]interface{}
	objrec, _ := json.Marshal(route)
	json.Unmarshal(objrec, &out)
	item.SetUnstructuredContent(out)

	return item, nil, nil
}
