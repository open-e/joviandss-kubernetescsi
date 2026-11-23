package driver

import (
	"github.com/sirupsen/logrus"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// What the two drivers share in their CSI lists: which two entries of one
// name are one object, and which entries a reply carries.

// sameVolume reports whether two entries of one name in a read of the
// volume list are one volume: the three fields identify the object and do
// not change with writes, as its sizes do. is_clone may be missing.
func sameVolume(a, b jrest.ResourceVolume) bool {
	return a.DefaultScsiId == b.DefaultScsiId && a.Guid == b.Guid && sameOptional(a.IsClone, b.IsClone)
}

// sameNASVolume always answers yes: an entry of the dataset list has no
// field that tells two datasets of one name apart, so a name that comes
// twice is dropped once.
func sameNASVolume(a, b jrest.ResourceNASVolume) bool { return true }

// sameSnapshot reports whether two entries of one name in a read of a
// snapshot list are one snapshot. Two snapshots of one name on two volumes
// differ in it.
func sameSnapshot(a, b jrest.ResourceSnapshotShort) bool {
	return a.Properties.Guid == b.Properties.Guid
}

func sameNASSnapshot(a, b jrest.ResourceNASSnapshotShort) bool {
	return a.Properties.Guid == b.Properties.Guid
}

// ownVolumeName reports whether the plugin made the name of a volume, and
// warns about one it did not make: the appliance's lists also hold objects
// that are not the plugin's, and a reply leaves them out.
func ownVolumeName(l *logrus.Entry, name string) bool {
	if _, err := NewVolumeDescFromVDS(name); err != nil {
		l.Warnf("Volume name has incompatible format %s", name)
		return false
	}
	return true
}

// ownSnapshotName does the same for a snapshot of the volume ld.
func ownSnapshotName(l *logrus.Entry, ld LunDesc, name string) bool {
	if _, err := NewSnapshotDescFromSDS(ld, name); err != nil {
		l.Warnf("Snapshot name has incompatible format %s", name)
		return false
	}
	return true
}

// ownSnapshotNames does the same for an entry of a list of all snapshots:
// the name of its parent, then its own.
func ownSnapshotNames(l *logrus.Entry, parent, name string) bool {
	vd, err := NewVolumeDescFromVDS(parent)
	if err != nil {
		l.Warnf("Volume name has incompatible format %s", parent)
		return false
	}
	return ownSnapshotName(l, vd, name)
}
