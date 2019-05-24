package main

import (
	"github.com/fusor/ocp-velero-plugin/velero-plugins/build"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/daemonset"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/deployment"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/deploymentconfig"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/imagestream"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/imagestreamtag"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/job"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/pod"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/pv"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/pvc"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/replicaset"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/replicationcontroller"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/route"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/statefulset"
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer().
		RegisterBackupItemAction("openshift.io/01-common-backup-plugin", newCommonBackupPlugin).
		RegisterRestoreItemAction("openshift.io/01-common-restore-plugin", newCommonRestorePlugin).
		RegisterBackupItemAction("openshift.io/02-pv-backup-plugin", newPVBackupPlugin).
		RegisterRestoreItemAction("openshift.io/03-pv-restore-plugin", newPVRestorePlugin).
		RegisterRestoreItemAction("openshift.io/03-pvc-restore-plugin", newPVCRestorePlugin).
		RegisterBackupItemAction("openshift.io/03-is-backup-plugin", newImageStreamBackupPlugin).
		RegisterRestoreItemAction("openshift.io/03-is-restore-plugin", newImageStreamRestorePlugin).
		RegisterRestoreItemAction("openshift.io/04-imagestreamtag-restore-plugin", newImageStreamTagRestorePlugin).
		RegisterRestoreItemAction("openshift.io/05-route-restore-plugin", newRouteRestorePlugin).
		RegisterRestoreItemAction("openshift.io/06-build-restore-plugin", newBuildRestorePlugin).
		RegisterRestoreItemAction("openshift.io/07-pod-restore-plugin", newPodRestorePlugin).
		RegisterBackupItemAction("openshift.io/08-deploymentconfig-backup-plugin", newDeploymentConfigBackupPlugin).
		RegisterRestoreItemAction("openshift.io/08-deploymentconfig-restore-plugin", newDeploymentConfigRestorePlugin).
		RegisterRestoreItemAction("openshift.io/09-replicationcontroller-restore-plugin", newReplicationControllerRestorePlugin).
		RegisterRestoreItemAction("openshift.io/10-job-restore-plugin", newJobRestorePlugin).
		RegisterRestoreItemAction("openshift.io/11-daemonset-restore-plugin", newDaemonSetRestorePlugin).
		RegisterRestoreItemAction("openshift.io/12-replicaset-restore-plugin", newReplicaSetRestorePlugin).
		RegisterRestoreItemAction("openshift.io/13-deployment-restore-plugin", newDeploymentRestorePlugin).
		RegisterRestoreItemAction("openshift.io/14-statefulset-restore-plugin", newStatefulSetRestorePlugin).
		Serve()
}

func newImageStreamBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &imagestream.BackupPlugin{Log: logger}, nil
}

func newImageStreamRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &imagestream.RestorePlugin{Log: logger}, nil
}

func newImageStreamTagRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &imagestreamtag.RestorePlugin{Log: logger}, nil
}

func newBuildRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &build.RestorePlugin{Log: logger}, nil
}

func newRouteRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &route.RestorePlugin{Log: logger}, nil
}

func newCommonBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &common.BackupPlugin{Log: logger}, nil
}

func newCommonRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &common.RestorePlugin{Log: logger}, nil
}

func newPodRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &pod.RestorePlugin{Log: logger}, nil
}

func newDeploymentConfigBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &deploymentconfig.BackupPlugin{Log: logger}, nil
}

func newDeploymentConfigRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &deploymentconfig.RestorePlugin{Log: logger}, nil
}

func newJobRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &job.RestorePlugin{Log: logger}, nil
}

func newReplicationControllerRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &replicationcontroller.RestorePlugin{Log: logger}, nil
}

func newDeploymentRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &deployment.RestorePlugin{Log: logger}, nil
}

func newReplicaSetRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &replicaset.RestorePlugin{Log: logger}, nil
}

func newDaemonSetRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &daemonset.RestorePlugin{Log: logger}, nil
}

func newStatefulSetRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &statefulset.RestorePlugin{Log: logger}, nil
}

func newPVBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &pv.BackupPlugin{Log: logger}, nil
}

func newPVRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &pv.RestorePlugin{Log: logger}, nil
}

func newPVCRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &pvc.RestorePlugin{Log: logger}, nil
}
