/*
Copyright (c) 2023 Open-E, Inc.
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
	"fmt"
	"os"

	//"github.com/open-e/JovianDSS-KubernetesCSI/pkg/joviandss"

	"github.com/spf13/cobra"
	cliPublishVolume "joviandss-kubernetescsi/cmd/controller/publishvolume"
	cli_common "joviandss-kubernetescsi/pkg/common"
)

var (
	volumeId           string
	volumeName         string
	volumeSize         string
	snapshotId         string
	nodeId             string
	volumeSizeRequired int64
	volumeSizeLimit    int64
	sourceVolumeName   string
	sourceSnapshotName string
	maxent             int32
	token              string
	readOnly           bool
)

// controllerCmd represents the controller command
var ControllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Set of commands associated with controller instance",
	Long: `Provides controller functionality including:
createVolume`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("controller called")
	},
}

func addSubCmds() {
	ControllerCmd.AddCommand(cliPublishVolume.PublishVolumeCmd)
	ControllerCmd.AddCommand(cliPublishVolume.UnpublishVolumeCmd)
}

func init() {
	cFlags := ControllerCmd.PersistentFlags()
	cFlags.StringVarP(
		&cli_common.ControllerConfigPath, "config", "c", "", "Path to controller config")

	if err := cobra.MarkFlagRequired(cFlags, "config"); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred:", err)
		os.Exit(1)
	}

	if err := cobra.MarkFlagFilename(cFlags, "config", "yml", "yaml"); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred:", err)
		os.Exit(1)
	}

	addSubCmds()
	// rootCmd.AddCommand(controllerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// controllerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// controllerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
