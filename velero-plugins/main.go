package main

import (
	"github.com/fusor/ocp-velero-plugin/velero-plugins/build"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/buildconfig"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/imagestream"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/route"
	veleroplugin "github.com/heptio/velero/pkg/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer().
		RegisterBackupItemAction("common-backup-plugin", newCommonBackupPlugin).
		RegisterRestoreItemAction("common-restore-plugin", newCommonRestorePlugin).
		RegisterBackupItemAction("is-backup-plugin", newImageStreamBackupPlugin).
		RegisterRestoreItemAction("is-restore-plugin", newImageStreamRestorePlugin).
		RegisterBackupItemAction("route-backup-plugin", newRouteBackupPlugin).
		RegisterRestoreItemAction("route-restore-plugin", newRouteRestorePlugin).
		RegisterRestoreItemAction("build-restore-plugin", newBuildRestorePlugin).
		Serve()
}

func newImageStreamBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &imagestream.BackupPlugin{Log: logger}, nil
}

func newImageStreamRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &imagestream.RestorePlugin{Log: logger}, nil
}

func newBuildConfigBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &buildconfig.BackupPlugin{Log: logger}, nil
}

func newBuildConfigRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &buildconfig.RestorePlugin{Log: logger}, nil
}

func newBuildRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &build.RestorePlugin{Log: logger}, nil
}

func newRouteBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &route.BackupPlugin{Log: logger}, nil
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
