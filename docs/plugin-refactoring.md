# ocp-velero-plugin Repo Split

Currently, the contents of ocp-velero-plugin includes a mixture of
actions that are relevant for general backup and restore of OpenShift
namespaces, and actions that only make sense in the context of a
controller-run migration. This document is an attempt to separate the
two purposes in preparation for splitting the repo into two.

## General backup/restore features (and supporting infrastructure):

* Record backup/restore cluster version and registry hostname.
* Don't restore pods that are controlled by higher-level entities (DeploymentConfig, etc.).
* Modify image references that point to the internal registry.
* Modify builds that point to the internal registry and update the builder-dockercfg secret.
* Modify subdomain on route to match dest cluster.

## Only needed for migration:

* ConfigureContainerSleep and plugins that deal with this
* ImageStream/ImageStreamTag image copy and "item skipping"
* PV/PVC actions
* Quiesce pods on final migration

## Post-split:

The general openshift velero plugin will be called
openshift-velero-plugin. The migration plugin will be
openshift-migration-plugin.

Once the repos are split, the migration-specific plugin image will
vendor the general plugin code as a dep (and register any general plugins that
will be used unchanged for migration). As a result, a velero
installation used for migration would only need to have the migration
plugin added, not the general plugin. Another approach would be to
keep them separate and require both plugins to be added, but this
reduces our flexibility to handle situations where we would want a
migration-specific plugin action registered *instead of* a general
plugin rather than *in addition to* it.

We should be able to remove the conditional statements that skip
plugin actions for non-migration use cases, since the migration plugin
won't be registered for general backup restore. This means that most
of the code like this won't be necessary anymore:
```
	if input.Restore.Annotations[common.MigrateCopyPhaseAnnotation] != "" {
```
We will still need conditionals where the action differs based on copy
vs. final migrations or copy vs. move behavior.

## File-by-file


* build
  * General
    * restore.go, restore_test.go
* clients
  * General
    * clients.go
* common:
  * General
    * backup.go  
    * restore.go
    * shared.go:
      * func getMetadataAndAnnotations
      * func GetServerVersion
      * func GetRegistryInfo
    * types.go:
      * type APIServerConfig
      * const BackupServerVersion
      * const RestoreServerVersion
      * const BackupRegistryHostname
      * const RestoreRegistryHostname
    * util.go
      * func ReplaceImageRefPrefix
      * func HasImageRefPrefix
      * type LocalImageReference
      * func ParseLocalImageReference
      * func SwapContainerIMageRefs
      * func GetSrcAndDestRegistryInfo
      * func GetOwnerReferences
  * Migration-specific:
      * const MigrationRegistry
      * const SwingPVAnnotation
      * const MigrateTypeAnnotation
      * const MigrateCopyPhaseAnnotation
      * const MigrateQuiesceAnnotation
      * const PodStageLabel
      * const ResticRestoreAnnotationPrefix
      * const ResticBackupAnnotation
    * util.go
      * func ConfigureContainerSleep
* daemonset:
  * General
    * restore.go
* deployment:
  * General
    * restore.go
  * Migration-specific:
    * backup.go (quiesce behavior)
* deploymentconfig:
  * General
    * restore.go
  * Migration-specific:
    * backup.go (quiesce behavior)
* imagestream:
  * Migration-specific:
    * backup.go, restore.go, shared.go (it's possible that we would want a general plugin
      for imagestream that's not used for migration, but nothing in these files should be used outside of migration)
* imagestreamtag:
  * Migration-specific:
    * restore.go   (it's possible that we would want a general plugin
      for imagestream that's not used for migration, but nothing in this file should be used outside of migration)
* job:
  * General
    * restore.go
* pod:
  * General
    * restore.go (swap image refs portion)
  * Migration-specific:
    * restore.go (ConfigureContainerSleep portion)
    * Note: here's an example where we need separate Execute action
      plugin for general and migration use cases if modifying
      image references goes into the general plugin
* pv:
  * Migration-specific:
    * backup.go
    * restore.go
* pvc:
  * Migration-specific:
    * restore.go
* replicaset:
  * General
    * restore.go
* replicationcontroller:
  * General
    * restore.go
* route:
  * General
    * restore.go, restore_test.go
* statefulset:
  * General
    * restore.go
* util/test:
  * General:
    * test.NewLogger
