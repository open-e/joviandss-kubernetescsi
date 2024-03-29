package node

import (
	"fmt"
	"os/exec"
	"crypto/sha256"
	"encoding/base64"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"joviandss-kubernetescsi/pkg/common"
	//"golang.org/x/net/context"
)

//var nodeId = ""

func GetNodeId(l *log.Entry) (string, error) {

	if len(common.NodeID) > 0 {
		l.Debugf("Node id identified %s", common.NodeID)

		return common.NodeID, nil
	}

	infostr := ""
	if out, err := exec.Command("hostname").Output(); err == nil {
		infostr = fmt.Sprintf("%s%s", infostr, out)
	}

	if out, err := exec.Command("cat", "/etc/machine-id").Output(); err == nil {
		infostr = fmt.Sprintf("%s%s", infostr, out)
	}

	if len(infostr) == 0 {
		return "", status.Errorf(codes.Internal, "Unable to identify node")
	}
	//l.Debugf("Node id %s", infostr)
	rawID := sha256.Sum256([]byte(infostr))
	common.NodeID = base64.StdEncoding.EncodeToString(rawID[:])

	//nodeId = string(rawID[:])

	return common.NodeID, nil
}
