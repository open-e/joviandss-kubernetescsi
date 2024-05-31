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
package cmd

import (
	"os"
	"github.com/open-e/joviandss-kubernetescsi/cmd/controller"
	"github.com/open-e/joviandss-kubernetescsi/cmd/node"
	"github.com/spf13/cobra"
	
	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "JovianDSS-CSI-CLI",
	Short: "Tool that allows user to use all driver functionality from command line",
	Long: `JovianDSS CSI CLI is a wrap arown JovianDSS CSI Plugin that allows user
	to debug and CSI plugin code and make actions as if this actions was initiated by Kubernetes
	CSI subsystem.

	This CLI does not communicate with Kubernetes.
	All user actions affects JovianDSS storage only.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	ConfigPath string = ""
)

func addSubCmds() {
	rootCmd.AddCommand(node.NodeCmd)
	rootCmd.AddCommand(controller.ControllerCmd)
}

func init() {
	var proto string
	rootCmd.PersistentFlags(). StringVarP(&proto, "prt", "", "iscsi", "Protocol type, can be iscsi or nfs")
	rootCmd.PersistentFlags().StringVarP(&jcom.LogLevel, "loglevel", "l", "INFO", "Level of logging")
	rootCmd.PersistentFlags().StringVarP(&jcom.LogPath, "logpath", "", "", "File to store log information")

	addSubCmds()
}
