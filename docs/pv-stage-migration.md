# Stage vs Migrate for persistent storage

There will be two phases to the migration that the controller will run. One
will be `stage`, and the other `migrate`. `stage` will be used to help copy the
persistent volume data over to the new cluster to reduce downtime. `stage` can
be run multiple times.

`migrate` will be used as the final cutover of the migration so that the
applications will be quiesced on the source cluster and launched on the new
cluster with the final state of the persistent volumes.

## Denoting phases with annotations

`stage` vs `migrate` backup/restores can be denoted using annotations on the
backup/restore CR. The controller will be responsible for annotating the CRs
with `openshit.io/stage-migration` or `openshift.io/migrate-migration`. The PV
backup plugin should function identically regardless of the annotations, but
restore will use these annotations to determine how PVs are being restored, and
whether or not the app will be migrated at the same time.

## Controlling backup/restore via controller

Since `stage` actions only care about persistent volumes, we need the
controller to be creating backups that only specify persistent volumes and
exclude all other resources. To do this, the controller should be creating
backups with the following spec for `stage` actions:
```yaml
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: foo
  namespace: velero
spec:
  includedResources:
  - 'persistentvolumes'
  excludedResources:
  - '*'
```
