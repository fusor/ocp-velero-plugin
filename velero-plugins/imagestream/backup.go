package imagestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	imagev1API "github.com/openshift/api/image/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/containers/image/copy"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"
	"github.com/containers/image/types"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/backup"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

type imagePolicyConfig struct {
	InternalRegistryHostname string `json:"internalRegistryHostname"`
}

type apiServerConfig struct {
	ImagePolicyConfig imagePolicyConfig `json:"imagePolicyConfig"`
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{
		IncludedResources: []string{"imagestreams"},
	}, nil
}

// Execute copies local registry images into migration registry
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.Log.Info("Hello from Imagestream backup plugin!!")

	im := imagev1API.ImageStream{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &im)
	p.Log.Info(fmt.Sprintf("image: %#v", im))
	annotations := im.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["openshift.io/imagestream-plugin"] = "1"

	internalRegistry, err := getRegistryInfo(p, backup.Annotations["openshift.io/backup-ocp-version"])
	if err != nil {
		return nil, nil, err
	}
	migrationRegistry := backup.Annotations["openshift.io/migration-registry"]
	if len(migrationRegistry) == 0 {
		return nil, nil, errors.New("migration registry not found for annotation \"openshift.io/migration\"")
	}
	// FIXME: do we need to compare registry to dockerRepo?
	//dockerRepo := im.Status.DockerImageRepository
	p.Log.Info(fmt.Sprintf("internal registry: %#v", internalRegistry))

	localImageCopied := false
	localImageCopiedByTag := false
	for _, tag := range im.Status.Tags {
		p.Log.Info(fmt.Sprintf("tag: %#v", tag.Tag))
		// FIXME: remove this comment once the below logic is implemented in the other plugins
		// in imagestreamtag restore plugin: restore if not local || has a tag reference
		// in imagestream backup plugin: copy all local images image (reverse order per tag):
		//                                1) if tag has null tag reference, copy to dest:tag
		//                                2) if there is a tag reference, copy to dest@sha
		// in imagestream restore plugin: copy all local images image (reverse order per tag):
		//                                1) if tag has null tag reference, copy to dest:tag
		//                                2) if there is a tag reference, copy to dest@sha
		//                                restore imagestream if no local images copied via dest tag
		specTag := findSpecTag(im.Spec.Tags, tag.Tag)
		copyToTag := true
		if specTag != nil && specTag.From != nil {
			p.Log.Info(fmt.Sprintf("image tagged: %s, %s", specTag.From.Kind, specTag.From.Name))
			// we have a tag.
			copyToTag = false
		}
		// Iterate over items in reverse order so most recently tagged is copied last
		for i := len(tag.Items) - 1; i >= 0; i-- {
			dockerImageReference := tag.Items[i].DockerImageReference
			localImage := strings.HasPrefix(dockerImageReference, internalRegistry)
			//p.Log.Info(fmt.Sprintf("image is local?: %t", localImage))
			if localImage {
				localImageCopied = true
				destTag := "@" + tag.Items[i].Image
				if copyToTag {
					localImageCopiedByTag = true
					destTag = ":" + tag.Tag
				}
				manifest, err := copyImage(p, fmt.Sprintf("docker://%s", dockerImageReference), fmt.Sprintf("docker://%s/%s/%s%s", migrationRegistry, im.Namespace, im.Name, destTag))
				if err != nil {
					return nil, nil, err
				}
				p.Log.Info(fmt.Sprintf("manifest of copied image: %s", manifest))
				p.Log.Info(fmt.Sprintf("copied from: docker://%s", dockerImageReference))
				p.Log.Info(fmt.Sprintf("copied to: docker://%s/%s/%s%s", migrationRegistry, im.Namespace, im.Name, destTag))
			}
		}
	}
	p.Log.Info(fmt.Sprintf("copied at least one local image: %t", localImageCopied))
	p.Log.Info(fmt.Sprintf("copied at least one local image by tag: %t", localImageCopiedByTag))

	im.Annotations = annotations

	var out map[string]interface{}
	objrec, _ := json.Marshal(im)
	json.Unmarshal(objrec, &out)
	item.SetUnstructuredContent(out)

	return item, nil, nil
}

func findSpecTag(tags []imagev1API.TagReference, name string) *imagev1API.TagReference {
	for _, tag := range tags {
		if tag.Name == name {
			return &tag
		}
	}
	return nil
}

func findStatusTag(tags []imagev1API.NamedTagEventList, name string) *imagev1API.NamedTagEventList {
	for _, tag := range tags {
		if tag.Tag == name {
			return &tag
		}
	}
	return nil
}

func getRegistryInfo(p *BackupPlugin, ocpVersion string) (string, error) {
	cClient, err := p.coreClient()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(ocpVersion, "3.") {
		registrySvc, err := cClient.Services("default").Get("docker-registry", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		internalRegistry := registrySvc.Spec.ClusterIP + ":" + strconv.Itoa(int(registrySvc.Spec.Ports[0].Port))
		return internalRegistry, nil
	} else if strings.HasPrefix(ocpVersion, "4.") {
		config, err := cClient.ConfigMaps("openshift-apiserver").Get("config", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		serverConfig := apiServerConfig{}
		err = json.Unmarshal([]byte(config.Data["config.yaml"]), &serverConfig)
		if err != nil {
			return "", err
		}
		internalRegistry := serverConfig.ImagePolicyConfig.InternalRegistryHostname
		if len(internalRegistry) == 0 {
			return "", errors.New("InternalRegistryHostname not found")
		}
		return internalRegistry, nil
	} else {
		return "", fmt.Errorf("OCP version %q not supported. Must be 3.x/4.x", ocpVersion)
	}
}

func (p *BackupPlugin) imageClient() (*imagev1.ImageV1Client, error) {
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

func (p *BackupPlugin) coreClient() (*corev1.CoreV1Client, error) {
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

func copyImage(p *BackupPlugin, src, dest string) (string, error) {
	policyContext, err := getPolicyContext()
	if err != nil {
		return "", fmt.Errorf("Error loading trust policy: %v", err)
	}
	defer policyContext.Destroy()

	srcRef, err := alltransports.ParseImageName(src)
	if err != nil {
		return "", fmt.Errorf("Invalid source name %s: %v", src, err)
	}
	destRef, err := alltransports.ParseImageName(dest)
	if err != nil {
		return "", fmt.Errorf("Invalid destination name %s: %v", dest, err)
	}
	sourceCtx, err := internalRegistrySystemContext()
	if err != nil {
		return "", err
	}
	destinationCtx, err := migrationRegistrySystemContext()
	if err != nil {
		return "", err
	}

	manifest, err := copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:      sourceCtx,
		DestinationCtx: destinationCtx,
	})
	return string(manifest), err
}

// getPolicyContext returns a *signature.PolicyContext based on opts.
func getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	return signature.NewPolicyContext(policy)
}

func internalRegistrySystemContext() (*types.SystemContext, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	ctx := &types.SystemContext{
		DockerDaemonInsecureSkipTLSVerify: true,
		DockerInsecureSkipTLSVerify:       types.OptionalBoolTrue,
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: "ignored",
			Password: config.BearerToken,
		},
	}
	return ctx, nil
}

func migrationRegistrySystemContext() (*types.SystemContext, error) {
	ctx := &types.SystemContext{
		DockerDaemonInsecureSkipTLSVerify: true,
		DockerInsecureSkipTLSVerify:       types.OptionalBoolTrue,
	}
	return ctx, nil
}
