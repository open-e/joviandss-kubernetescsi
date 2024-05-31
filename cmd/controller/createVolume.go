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

package controller

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	//humanize "github.com/dustin/go-humanize"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"

	"github.com/spf13/cobra"
)

var (
//volumeName string
//volumeSize string

//volumeSizeRequired string
//volumeSizeLimit string
)

func createVolume(cmd *cobra.Command, args []string) {
	logrus.Debug("create volume")

	var cfg common.JovianDSSCfg

	var cp controller.ControllerPlugin

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.CreateVolumeRequest
	var ctx context.Context = common.GetContext("create_volume")

	var bytes uint64 = 0

	var cr csi.CapacityRange
	cr.RequiredBytes = int64(bytes)
	req.CapacityRange = &cr

	req.Name = volumeName
	var supportedVolumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
		// VolumeCapability_AccessMode_UNKNOWN,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		// VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		// VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		// VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}

	req.VolumeCapabilities = controller.GetVolumeCapability(supportedVolumeCapabilities)

	if len(sourceSnapshotName) > 0 {
		req.VolumeContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: sourceSnapshotName,
				},
			},
		}
	}

	if len(sourceVolumeName) > 0 {
		req.VolumeContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Volume{
				Volume: &csi.VolumeContentSource_VolumeSource{
					VolumeId: sourceVolumeName,
				},
			},
		}
	}

	if volumeSizeLimit != 0 {
		req.CapacityRange = &csi.CapacityRange{
			LimitBytes: volumeSizeLimit,
		}

	}

	if volumeSizeRequired != 0 {
		req.CapacityRange = &csi.CapacityRange{
			RequiredBytes: int64(volumeSizeRequired),
		}
	}
	resp, err := cp.CreateVolume(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}

// createVolumeCmd represents the createvolume command
var createVolumeCmd = &cobra.Command{
	Use:   "createVolume",
	Short: "Create volume",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: createVolume,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	createVolumeCmd.Flags().StringVarP(&volumeName, "name", "n", "", "Name of volume to create")
	//createVolumeCmd.Flags().StringVarP(&volumeSize,"size",	"s",		"", "Size of volume to create")
	createVolumeCmd.Flags().Int64VarP(&volumeSizeRequired, "srq", "", 0, "Required size of volume to create")
	createVolumeCmd.Flags().Int64VarP(&volumeSizeLimit, "slm", "", 0, "Limit size of volume to create")
	createVolumeCmd.Flags().StringVarP(&sourceVolumeName, "volume", "", "", "Name of source volume to use")
	createVolumeCmd.Flags().StringVarP(&sourceSnapshotName, "snapshot", "", "", "Name of source snapshot to use")

	if err := createVolumeCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}

	ControllerCmd.AddCommand(createVolumeCmd)
}
