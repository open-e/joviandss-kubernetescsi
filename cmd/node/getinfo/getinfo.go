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

package getinfo

import (
	"context"
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	//humanize "github.com/dustin/go-humanize"

	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_node "joviandss-kubernetescsi/pkg/node"

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

func getInfo(cmd *cobra.Command, args []string) {

	// var np csi_node.NodePlugin

	logger, err := cli_common.GetLogger(csi_common.LogLevel, csi_common.LogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init loging because:", err.Error())
		os.Exit(1)
	}
	l := log.NewEntry(logger)
	l.Debug("publish volume")

	np, err := csi_node.GetNodePlugin(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to init node plugin:", err.Error())
		os.Exit(1)
	}
	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.NodeGetInfoRequest
	var ctx context.Context = common.GetContext("node_getinfo")

	resp, err := np.NodeGetInfo(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
}

// createVolumeCmd represents the createvolume command
var NodeGetInfoCmd = &cobra.Command{
	Use:   "getinfo",
	Short: "gets node info",
	Long: `Gets node information that is available by plugin.

	That makes volume available for attachment`,
	Run: getInfo,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	//ControllerCmd.AddCommand(publishVolumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
