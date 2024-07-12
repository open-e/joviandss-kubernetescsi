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

package driverapi

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	create_volume "github.com/open-e/joviandss-kubernetescsi/cmd/driverapi/createvolume"
	cli_common "github.com/open-e/joviandss-kubernetescsi/pkg/common"
	// cliStageVolume "github.com/open-e/joviandss-kubernetescsi/cmd/node/stagevolume"
)

var DriverCmd = &cobra.Command{
	Use:   "driver",
	Short: "Comand line interface to driver API",
	Long:  `That sub command provides interface to call driver API using CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("driver API called")
	},
}

func addSubCmds() {
	DriverCmd.AddCommand(create_volume.CreateVolumeCmd)
}

func init() {
	cFlags := DriverCmd.PersistentFlags()
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
	// rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
