package common

import (
	"fmt"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
)

// RestorePlugin is a restore item action plugin for Heptio Ark.
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to everything.
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute sets a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Info("Hello from common restore plugin!!")

	metadata, annotations, err := getMetadataAndAnnotations(input.Item)
	if err != nil {
		return nil, err
	}

	version, err := getVersion()

	annotations[RestoreServerVersion] = fmt.Sprintf("%v.%v", version.Major, version.Minor)
	registryHostname, err := getRegistryInfo(version.Major, version.Minor)
	if err != nil {
		return nil, err
	}
	annotations[RestoreRegistryHostname] = registryHostname
	metadata.SetAnnotations(annotations)

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}
