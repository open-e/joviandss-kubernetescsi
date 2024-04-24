/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package controller

import (
	"context"
	"fmt"

	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"
	"github.com/open-e/joviandss-kubernetescsi/pkg/common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteVolumeCmd represents the deleteVolume command
var deleteVolumeCmd = &cobra.Command{
	Use:   "deleteVolume",
	Short: "Delete specified volume",
	Long:  ``,
	Run:   deleteVolume,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("deleteVolume called")
	// },
}

func deleteVolume(cmd *cobra.Command, args []string) {

	logrus.Debug("delete volumes")
	var cfg common.JovianDSSCfg
	// controller.ControllerCfg
	// var cp csi_controller.ControllerPlugin

	var cp controller.ControllerPlugin

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

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

	if err := deleteVolumeCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}

	ControllerCmd.AddCommand(deleteVolumeCmd)
}
