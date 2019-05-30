package pv

import (
	"encoding/json"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
	corev1API "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"persistentvolumes"},
	}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.Log.Info("Hello from PV backup plugin!!")

	// Convert to PV
	backupPV := corev1API.PersistentVolume{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &backupPV)

	client, err := clients.CoreClient()
	if err != nil {
		return nil, nil, err
	}
	// Get and update PVC on the running cluster to use a retain policy
	// Validate PVC wasn't deleted by getting the object from the cluster
	pv, err := client.PersistentVolumes().Get(backupPV.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	// Set reclaimPolicy to retain if swinging PV
	if pv.Annotations[common.MigrateTypeAnnotation] == "move" {
		pv.Spec.PersistentVolumeReclaimPolicy = corev1API.PersistentVolumeReclaimRetain
	}

	// Update PV
	pv, err = client.PersistentVolumes().Update(pv)
	if err != nil {
		return nil, nil, err
	}

	return item, nil, nil
}
