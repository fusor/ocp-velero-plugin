package route

import (
	"encoding/json"
	"strings"

	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	routev1API "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type routingConfig struct {
	Subdomain string `json:"subdomain"`
}

type imagePolicyConfig struct {
	InternalRegistryHostname string `json:"internalRegistryHostname"`
}

type ApiServerConfig struct {
	ImagePolicyConfig imagePolicyConfig `json:"imagePolicyConfig"`
	RoutingConfig     routingConfig     `json:"routingConfig"`
}

// MyRestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything
func (p *RestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"routes"},
	}, nil
}

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

	client, err := p.coreClient()
	if err != nil {
		return nil, nil, err
	}
	config, err := client.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	serverConfig := ApiServerConfig{}
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
