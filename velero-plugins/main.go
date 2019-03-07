/*
Copyright 2017 the Heptio Ark contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/fusor/ocp-velero-plugin/velero-plugins/buildconfig"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/imagestream"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/route"
	veleroplugin "github.com/heptio/velero/pkg/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer(veleroplugin.NewLogger()).
		RegisterBackupItemAction("is-backup-plugin", newImageStreamBackupPlugin).
		RegisterRestoreItemAction("is-restore-plugin", newImageStreamRestorePlugin).
		RegisterBackupItemAction("route-backup-plugin", newRouteBackupPlugin).
		RegisterRestoreItemAction("route-restore-plugin", newRouteRestorePlugin).
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

func newRouteBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &route.BackupPlugin{Log: logger}, nil
}

func newRouteRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &route.RestorePlugin{Log: logger}, nil
}
