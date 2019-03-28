package imagestream

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	imagev1API "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
        "github.com/heptio/velero/pkg/plugin/velero"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"imagestreams"},
	}, nil
}

// Execute copies local registry images into migration registry
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
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

	internalRegistry := annotations[common.BackupRegistryHostname]
	if len(internalRegistry) == 0 {
		return nil, nil, errors.New("migration registry not found for annotation \"openshift.io/backup-registry-hostname\"")
	}
	migrationRegistry := backup.Annotations[common.MigrationRegistry]
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
				manifest, err := copyImageBackup(fmt.Sprintf("docker://%s", dockerImageReference), fmt.Sprintf("docker://%s/%s/%s%s", migrationRegistry, im.Namespace, im.Name, destTag))
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

func copyImageBackup(src, dest string) (string, error) {
	sourceCtx, err := internalRegistrySystemContext()
	if err != nil {
		return "", err
	}
	destinationCtx, err := migrationRegistrySystemContext()
	if err != nil {
		return "", err
	}
	return copyImage(src, dest, sourceCtx, destinationCtx)
}
