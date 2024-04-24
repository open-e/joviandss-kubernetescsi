/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package publishvolume

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"

	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"

	"github.com/spf13/cobra"
)

var (
	readOnly bool
)

func publishVolume(cmd *cobra.Command, args []string) {
	logrus.Debug("publish volume")

	var cfg common.JovianDSSCfg

	var cp controller.ControllerPlugin

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

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
