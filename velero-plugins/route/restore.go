package route

import (
	"encoding/json"
	"strings"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	"github.com/heptio/velero/pkg/plugin/velero"
	routev1API "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to everything
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"routes"},
	}, nil
}

// Execute fixes the route path on restore to use the target cluster's domain name
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Info("Hello from Route RestorePlugin!")
	route := routev1API.Route{}
	itemMarshal, _ := json.Marshal(input.Item)
	json.Unmarshal(itemMarshal, &route)

	client, err := clients.CoreClient()
	if err != nil {
		return nil, err
	}
	config, err := client.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	serverConfig := common.APIServerConfig{}
	err = json.Unmarshal([]byte(config.Data["config.yaml"]), &serverConfig)
	if err != nil {
		return nil, err
	}

	subdomain := serverConfig.RoutingConfig.Subdomain

	output := replaceSubdomain(input.Item, &route, subdomain)

	return output, nil
}

func replaceSubdomain(item runtime.Unstructured, route *routev1API.Route, subdomain string) *velero.RestoreItemActionExecuteOutput {
	host := route.Spec.Host
	name := strings.Split(host, ".")[0]
	newHost := name + "." + subdomain
	route.Spec.Host = newHost

	var out map[string]interface{}
	objrec, _ := json.Marshal(route)
	json.Unmarshal(objrec, &out)
	item.SetUnstructuredContent(out)

	return velero.NewRestoreItemActionExecuteOutput(item)
}
