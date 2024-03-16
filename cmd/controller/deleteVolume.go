/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package controller

import (
	"context"
	"fmt"

	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_controller "joviandss-kubernetescsi/pkg/controller"
	cli_common "joviandss-kubernetescsi/pkg/common"

	// csi_rest "joviandss-kubernetescsi/pkg/rest"

	"joviandss-kubernetescsi/pkg/common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteVolumeCmd represents the deleteVolume command
var deleteVolumeCmd = &cobra.Command{
	Use:   "deleteVolume",
	Short: "Delete specified volume",
	Long: ``,
	Run: deleteVolume,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("deleteVolume called")
	// },
}



func deleteVolume(cmd *cobra.Command, args []string) {

	logrus.Debug("delete volumes")
	var cfg csi_common.JovianDSSCfg
	// controller.ControllerCfg
	// var cp csi_controller.ControllerPlugin

	var cp csi_controller.ControllerPlugin


	if err := csi_common.SetupConfig(cli_common.ControllerConfigPath, &cfg) ; err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	csi_controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.DeleteVolumeRequest
	var ctx context.Context = common.GetContext("delete_volume")
	req.VolumeId = volumeId
	_, err := cp.DeleteVolume(ctx, &req)

	if err != nil {
		logrus.Errorln("delete volume failes, error is ", err.Error())
	}

}

func init() {
	deleteVolumeCmd.Flags().StringVarP(&volumeId, "name", "n", "", "Name of volume to delete")

	if err:= deleteVolumeCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}

	ControllerCmd.AddCommand(deleteVolumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteVolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteVolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
