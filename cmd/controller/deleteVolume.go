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

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"

	// csi_rest "joviandss-kubernetescsi/pkg/rest"

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
	var cfg jcom.JovianDSSCfg
	// controller.ControllerCfg
	// var cp csi_controller.ControllerPlugin

	var cp controller.ControllerPlugin

	if err := jcom.SetupConfig(jcom.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

	var req csi.DeleteVolumeRequest
	var ctx context.Context = jcom.GetContext("delete_volume")
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
