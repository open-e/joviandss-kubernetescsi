/*
Copyright (c) 2024 Open-E, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package getinfo

import (
	"context"
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	//humanize "github.com/dustin/go-humanize"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	"github.com/open-e/joviandss-kubernetescsi/pkg/node"

	"github.com/spf13/cobra"
)

var (
	volume_id           string
	publish_context     map[string]string
	staging_target_path string
	volume_capabilty    []string
	secrets             map[string]string

	//volumeSize string

	//volumeSizeRequired string
	//volumeSizeLimit string
)

func stageVolume(cmd *cobra.Command, args []string) {

	// var np csi_node.NodePlugin

	logger, err := common.GetLogger(common.LogLevel, common.LogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init loging because:", err.Error())
		os.Exit(1)
	}
	l := log.NewEntry(logger)
	l.Debug("stage volume command")

	np, err := node.GetNodePlugin(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init node plugin:", err.Error())
		os.Exit(1)
	}
	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.NodeStageVolumeRequest
	var ctx context.Context = common.GetContext("node_stagevolume")

	mountVolume := csi.VolumeCapability_MountVolume{
		FsType:     "ext4",
		MountFlags: []string{"rw"},
	}

	// Define the access mode
	accessMode := csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}

	// Combine access mode and access type into a VolumeCapability
	volumeCapability := csi.VolumeCapability{
		AccessMode: &accessMode,
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &mountVolume,
		},
	}

	req.VolumeId = volume_id

	if len(publish_context) > 0 {
		req.PublishContext = publish_context
	}

	req.StagingTargetPath = staging_target_path

	req.VolumeCapability = &volumeCapability

	if len(secrets) > 0 {
		req.Secrets = secrets
	}

	resp, err := np.NodeStageVolume(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Request completed with response %+v\n", resp)
}

var NodeStageVolumeCmd = &cobra.Command{
	Use:   "stagevolume",
	Short: "StageVolume trigers stagevolume procedure",
	Long: `StageVolume emulates NodeStageVolume grpc Request from CO.

	https://github.com/container-storage-interface/spec/blob/master/spec.md#nodestagevolume`,
	Run: stageVolume,
}

func init() {

	NodeStageVolumeCmd.Flags().StringVarP(&volume_id, "volume_id", "i", "", "The ID of the volume to publish. This field is REQUIRED.")
	NodeStageVolumeCmd.Flags().StringToStringVarP(&publish_context, "publish_context", "c", map[string]string{}, "Context provided by controller after running PublishVolume, Optional.")

	staging_target_path_desc := `The path to which the volume MAY be staged. It MUST be an
	absolute path in the root filesystem of the process serving this
	request, and MUST be a directory. The CO SHALL ensure that there
	is only one 'staging_target_path' per volume. The CO SHALL ensure
	that the path is directory and that the process serving the
	request has 'read' and 'write' permission to that directory. The
	CO SHALL be responsible for creating the directory if it does not
	exist.`

	NodeStageVolumeCmd.Flags().StringVarP(&staging_target_path, "staging_target_path", "p", "", staging_target_path_desc)

	// TODO: implement volume capability
	// Volume capability describing how the CO intends to use this volume.
	// SP MUST ensure the CO can use the staged volume as described.
	// Otherwise SP MUST return the appropriate gRPC error code.
	// This is a REQUIRED field.
	// VolumeCapability volume_capability = 4;

	// Secrets required by plugin to complete node stage volume request.
	// This field is OPTIONAL. Refer to the `Secrets Requirements`
	// section on how to use this field.
	NodeStageVolumeCmd.Flags().StringToStringVarP(&secrets, "secrets", "s", map[string]string{}, "Secrets required by plugin to complete node stage volume request, Optional.")

}
