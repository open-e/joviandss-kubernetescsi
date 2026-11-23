package controller

import (
	"context"
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jdrvr "github.com/open-e/joviandss-kubernetescsi/pkg/driver"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// volumeDriver stands for a driver in the calls that read one volume.
type volumeDriver struct {
	jdrvr.CSIDriver

	volume *jdrvr.GeneralVolume
	err    jrest.RestError
	reads  int
}

func (d *volumeDriver) GetVolume(ctx context.Context, pool string, vd *jdrvr.VolumeDesc) (*jdrvr.GeneralVolume, jrest.RestError) {
	d.reads++
	return d.volume, d.err
}

// The failures of the first read of a volume that are not "not found": an
// appliance that cannot be reached, a certificate that does not verify, a
// failed login, a reply the plugin does not know.
var readFailures = map[string]jrest.RestError{
	"a certificate that does not verify": jrest.GetError(jrest.RestErrorUnableToConnect, "certificate of the appliance not verified: x509: unknown authority; set cafile to the certificate or its CA, or set insecure"),
	"a request that times out":           jrest.GetError(jrest.RestErrorRequestTimeout, "request timed out"),
	"a failed login":                     jrest.GetError(jrest.RestErrorAuthFailed, "authentication failed, status 401"),
	"a missing pool":                     jrest.GetError(jrest.RestErrorResourceDNEPool, "Pool Pool-0 not found."),
	"a reply the plugin does not know":   jrest.GetError(jrest.RestErrorFailureUnknown, "status 500"),
}

// VolumeComply reads a volume only when it has one. A read that fails with
// anything but "not found" is an error with its cause, never a panic, and
// "not found" stays NotFound.
func TestVolumeComplyWhenTheVolumeCannotBeRead(t *testing.T) {
	l := log.New()
	l.SetOutput(discardLog{})
	ctx := jcom.WithLogger(context.Background(), l.WithField("test", "VolumeComply"))
	vd, err := jdrvr.NewVolumeDescFromName("pvc-1")
	if err != nil {
		t.Fatal(err)
	}
	for what, failure := range readFailures {
		d := &volumeDriver{err: failure}
		cp := &ControllerPlugin{l: l, d: d, pool: "Pool-0"}
		size, err := cp.VolumeComply(ctx, vd, &csi.CapacityRange{RequiredBytes: 1 << 30}, nil)
		if size != nil || status.Code(err) != codes.Internal || !strings.Contains(err.Error(), failure.Error()) {
			t.Errorf("%s: size %v, error %v; want Internal with the cause", what, size, err)
		}
	}

	d := &volumeDriver{err: jrest.GetError(jrest.RestErrorResourceDNE, "no such volume")}
	cp := &ControllerPlugin{l: l, d: d, pool: "Pool-0"}
	if _, err := cp.VolumeComply(ctx, vd, nil, nil); status.Code(err) != codes.NotFound {
		t.Errorf("a volume that does not exist: %v, want NotFound", err)
	}

	// a driver that returns neither a volume nor an error
	d = &volumeDriver{}
	cp = &ControllerPlugin{l: l, d: d, pool: "Pool-0"}
	if _, err := cp.VolumeComply(ctx, vd, nil, nil); status.Code(err) != codes.Internal {
		t.Errorf("no volume and no error: %v, want Internal", err)
	}

	// a volume that is there and fits
	d = &volumeDriver{volume: &jdrvr.GeneralVolume{Name: "vp_pvc-1", Size: 1 << 30}}
	cp = &ControllerPlugin{l: l, d: d, pool: "Pool-0"}
	size, err := cp.VolumeComply(ctx, vd, &csi.CapacityRange{RequiredBytes: 1 << 30}, nil)
	if status.Code(err) != codes.OK || size == nil || *size != 1<<30 {
		t.Errorf("a volume that fits: size %v, error %v", size, err)
	}
}

// CreateVolume answers such a failure with an error that carries its
// cause: the controller does not crash, and nothing is created.
func TestCreateVolumeWhenTheVolumeCannotBeRead(t *testing.T) {
	l := log.New()
	l.SetOutput(discardLog{})
	req := &csi.CreateVolumeRequest{
		Name:               "pvc-1",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: 1 << 30},
		VolumeCapabilities: []*csi.VolumeCapability{{}},
	}
	for what, failure := range readFailures {
		d := &volumeDriver{err: failure}
		cp := &ControllerPlugin{l: l, d: d, pool: "Pool-0"}
		rsp, err := cp.CreateVolume(context.Background(), req)
		if rsp != nil || err == nil || status.Code(err) == codes.OK || !strings.Contains(err.Error(), failure.Error()) {
			t.Errorf("%s: response %v, error %v; want an error with the cause", what, rsp, err)
		}
		if d.reads != 1 {
			t.Errorf("%s: the volume was read %d times", what, d.reads)
		}
	}
}

type discardLog struct{}

func (discardLog) Write(p []byte) (int, error) { return len(p), nil }
