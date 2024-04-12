/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package controller

import (
	"context"
	"fmt"

	cli_common "joviandss-kubernetescsi/pkg/common"
	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_controller "joviandss-kubernetescsi/pkg/controller"

	// csi_rest "joviandss-kubernetescsi/pkg/rest"

	"joviandss-kubernetescsi/pkg/common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listVolumesCmd represents the listVolumes command
var listVolumesCmd = &cobra.Command{
	Use:   "listVolumes",
	Short: "Lists volumes for specific config",
	Long: `listVolumes give list of all volumes on the given pool
can do for all in one call or broken into sections`,
	Run: listVolumes,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("listVolumes called")
	// },
}

func listVolumes(cmd *cobra.Command, args []string) {

	logrus.Debug("list volumes")
	var cfg csi_common.JovianDSSCfg
	// controller.ControllerCfg
	// var cp csi_controller.ControllerPlugin

	var cp csi_controller.ControllerPlugin

	if err := csi_common.SetupConfig(cli_common.ControllerConfigPath, &cfg); err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	csi_controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.ListVolumesRequest
	var ctx context.Context = common.GetContext("list_volume")
	resp, err := cp.ListVolumes(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Got resp of size %d %+v\n", len(resp.Entries), resp)
	for i := 0; i < len(resp.Entries); i++ {
		fmt.Printf("volume %s\n", resp.Entries[i].Volume.VolumeId)
	}

	// var vols []csi_rest.Volume
	// var rEndpoint csi_rest.RestEndpoint
	// csi_rest.SetupEndpoint(&rEndpoint, &cfg.RestEndpointCfg)

	// if err := rEndpoint.ListVolumes("Pool-0", &vols) ; err != nil {
	// 	panic(err)
	// }

	// for i:=0 ; i < len(vols) ; i++ {
	// 	fmt.Printf("volume %s\n", vols[i].Name)
	// }
	//csi_rest.CreateV lume(
	//if err := csi_controller.GetConfig(ControllerConfigPath, &controllerCfg); err != nil {
	//	panic(err)
	//}
	//l := csi_common.GetLogger(cfg.LLevel, cfg.LPath)

	//if err := csi_controller.GetControllerPlugin(&cp, &cfg, l); err != nil {
	//		log.Fatalf("Unable to init controller: %v", err)
	//}
}

func init() {
	// rootCmd.AddCommand(listVolumesCmd)

	// createVolumeCmd.Flags().StringVarP(&volumeName,"name", "n", "", "Name of volume to create")

	// if err:= createVolumeCmd.MarkFlagRequired("name"); err != nil {
	// 	fmt.Println(err)
	// }

	ControllerCmd.AddCommand(listVolumesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listVolumesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listVolumesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
