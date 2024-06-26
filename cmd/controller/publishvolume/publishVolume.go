/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package publishvolume

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	//humanize "github.com/dustin/go-humanize"

	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_controller "joviandss-kubernetescsi/pkg/controller"

	cli_common "joviandss-kubernetescsi/pkg/common"

	"joviandss-kubernetescsi/pkg/common"

	"github.com/spf13/cobra"
)

var (
	readOnly bool
	//volumeSize string

	//volumeSizeRequired string
	//volumeSizeLimit string
)

func publishVolume(cmd *cobra.Command, args []string) {
	logrus.Debug("publish volume")

	var cfg csi_common.JovianDSSCfg

	var cp csi_controller.ControllerPlugin

	if err := csi_common.SetupConfig(cli_common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	csi_controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.ControllerPublishVolumeRequest
	var ctx context.Context = common.GetContext("publish_volume")

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

	req.VolumeCapability = &volumeCapability

	req.VolumeId = volumeId

	resp, err := cp.ControllerPublishVolume(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}

// createVolumeCmd represents the createvolume command
var PublishVolumeCmd = &cobra.Command{
	Use:   "publishVolume",
	Short: "publish volume",
	Long: `Creates apropriate target and attaches volume to it.

	That makes volume available for attachment`,
	Run: publishVolume,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	PublishVolumeCmd.Flags().StringVarP(&volumeId, "id", "i", "", "Id of volume to publish")
	PublishVolumeCmd.Flags().StringVarP(&nodeId, "nodeid", "n", "", "Id of node that volume will be published on")
	PublishVolumeCmd.Flags().BoolVarP(&readOnly, "readonly", "r", false, "Should volume be readonly")

	//ControllerCmd.AddCommand(publishVolumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
