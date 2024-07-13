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

package createvolume

import (
	"context"
	"fmt"
	"os"

	// "github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	jdrvr "github.com/open-e/joviandss-kubernetescsi/pkg/driver"

	"github.com/spf13/cobra"
)

var (
	volumeName string
	volumeSize int64

	sourceVolumeName   string
	sourceSnapshotName string
	volumeSizeRequired int64
	volumeSizeLimit    int64
)

var readOnly bool

func createVolume(cmd *cobra.Command, args []string) {
	// var np csi_node.NodePlugin

	var cfg common.JovianDSSCfg

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}

	logger, err := common.GetLogger(common.LogLevel, common.LogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init loging because:", err.Error())
		os.Exit(1)
	}
	l := log.NewEntry(logger)
	l.Debug("Create NFS share")

	d, err := jdrvr.NewJovianDSSCSINFSDriver(&cfg.RestEndpointCfg, l)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init driver:", err.Error())
		os.Exit(1)
	}

	var ctx context.Context = common.GetContext("driver_new_nas_volume")

	ctx = jcom.WithLogger(ctx, l)

	var nvid *jdrvr.VolumeDesc
	if nvid, err = jdrvr.NewVolumeDescFromName(volumeName); err != nil {
		panic(err)
	}

	if err = d.CreateVolume(ctx, "Pool-0", nvid, volumeSizeRequired, volumeSizeRequired); err != nil {
		panic(err)
	}
}

// createVolumeCmd represents the createvolume command
var CreateVolumeCmd = &cobra.Command{
	Use:   "createvolume",
	Short: "create nfs volume",
	Long: `Takes volume properties and creates new NFS volume on given pool.

	`,
	Run: createVolume,
}

func init() {
	// CreateVolumeCmd.Flags().StringVarP(&volumeName, "name", "n", "", "Name of NFS share create")
	// CreateVolumeCmd.Flags().Int64VarP(&volumeSize, "size", "s", 0, "Size of volume to create")

	CreateVolumeCmd.Flags().StringVarP(&volumeName, "name", "n", "", "Name of volume to create")
	CreateVolumeCmd.Flags().Int64VarP(&volumeSizeRequired, "sreq", "r", 0, "Required size of volume to create")
	CreateVolumeCmd.Flags().Int64VarP(&volumeSizeLimit, "slim", "m", 0, "Limit size of volume to create")
	CreateVolumeCmd.Flags().StringVarP(&sourceVolumeName, "volume", "v", "", "Name of source volume to use")
	CreateVolumeCmd.Flags().StringVarP(&sourceSnapshotName, "snapshot", "s", "", "Name of source snapshot to use")

	if err := CreateVolumeCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}
	if err := CreateVolumeCmd.MarkFlagRequired("size"); err != nil {
		fmt.Println(err)
	}

	// ControllerCmd.AddCommand(publishVolumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
