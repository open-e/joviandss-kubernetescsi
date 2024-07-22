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

	"github.com/open-e/joviandss-kubernetescsi/pkg/common"
	"github.com/open-e/joviandss-kubernetescsi/pkg/controller"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteVolumeCmd represents the deleteVolume command
var deleteSnapshotCmd = &cobra.Command{
	Use:   "deleteSnapshot",
	Short: "Delete specified snapshot",
	Long:  ``,
	Run:   deleteSnapshot,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("deleteVolume called")
	// },
}

func deleteSnapshot(cmd *cobra.Command, args []string) {
	logrus.Debug("delete snapshot")
	var cfg common.JovianDSSCfg
	// controller.ControllerCfg
	// var cp csi_controller.ControllerPlugin

	var cp controller.ControllerPlugin

	if err := common.SetupConfig(common.ControllerConfigPath, &cfg); err != nil {
		panic(err)
	}
	controller.SetupControllerPlugin(&cp, &cfg)

	var req csi.DeleteSnapshotRequest
	var ctx context.Context = common.GetContext("delete_snapshot")
	req.SnapshotId = snapshotId
	_, err := cp.DeleteSnapshot(ctx, &req)
	if err != nil {
		logrus.Errorln("delete snapshot failes, error is ", err.Error())
	}
}

func init() {
	deleteSnapshotCmd.Flags().StringVarP(&snapshotId, "name", "n", "", "Name of snapshot to delete")

	if err := deleteSnapshotCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}

	ControllerCmd.AddCommand(deleteSnapshotCmd)
}
