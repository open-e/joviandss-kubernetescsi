/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package publishvolume

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"

	"github.com/spf13/cobra"
)

var (
	volumeId string
	nodeId   string
)

func unpublishVolume(cmd *cobra.Command, args []string) {
	logrus.Debug("unpublish volume")

	var cfg common.JovianDSSCfg

	var cp controller.ControllerPlugin

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.ControllerUnpublishVolumeRequest
	var ctx context.Context = common.GetContext("unpublish_volume")

	req.VolumeId = volumeId
	req.NodeId = nodeId

	resp, err := cp.ControllerUnpublishVolume(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}

// createVolumeCmd represents the createvolume command
var UnpublishVolumeCmd = &cobra.Command{
	Use:   "unpublishVolume",
	Short: "Unpublish volume",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: unpublishVolume,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	UnpublishVolumeCmd.Flags().StringVarP(&volumeId, "volumeid", "i", "", "Name of volume to unstage")
	UnpublishVolumeCmd.Flags().StringVarP(&nodeId, "nodeid", "n", "", "Id of node that volume will be unstaged from")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
