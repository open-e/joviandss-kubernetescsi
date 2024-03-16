/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package controller

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"

	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_controller "joviandss-kubernetescsi/pkg/controller"
	
	cli_common "joviandss-kubernetescsi/pkg/common"
	
	"joviandss-kubernetescsi/pkg/common"

	"github.com/spf13/cobra"
)

var (
	snapshotName string
)

func createSnaposhot(cmd *cobra.Command, args []string) {
	logrus.Debug("create snapshot")
	
	var cfg csi_common.JovianDSSCfg

	var cp csi_controller.ControllerPlugin

	if err := csi_common.SetupConfig(cli_common.ControllerConfigPath, &cfg) ; err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	csi_controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.CreateSnapshotRequest
	var ctx context.Context = common.GetContext("create_snapshot")

	req.Name = snapshotName
	req.SourceVolumeId = volumeId

	resp, err := cp.CreateSnapshot(ctx, &req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", resp)
	//for i:=0 ; i < len(resp.Entries) ; i++ {
	//	fmt.Printf("volume %s\n",:23
	//resp.Entries[i].Volume.VolumeId)
	//}

	// var cfg csi_common.JovianDSSCfg
	// // controller.ControllerCfg
	// // var cp csi_controller.ControllerPlugin

	// if err := csi_common.SetupConfig(ControllerConfigPath, &cfg) ; err != nil {
	// 	// GetConfig(ControllerConfigPath, &controllerCfg)
	// 	panic(err)
	// }

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}
	// var rEndpoint csi_rest.RestEndpoint
	// csi_rest.SetupEndpoint(&rEndpoint, &cfg.RestEndpointCfg)

	// if err := rEndpoint.CreateVolume("Pool-0", vol) ; err != nil {
	// 	panic(err)
	// }
	//csi_rest.CreateVolume(
	//if err := csi_controller.GetConfig(ControllerConfigPath, &controllerCfg); err != nil {
	//	panic(err)
	//}
	//l := csi_common.GetLogger(cfg.LLevel, cfg.LPath)

	//if err := csi_controller.GetControllerPlugin(&cp, &cfg, l); err != nil {
	//		log.Fatalf("Unable to init controller: %v", err)
	//}
}

// createVolumeCmd represents the createvolume command
var createSnapshotCmd = &cobra.Command{
	Use:   "createSnapshot",
	Short: "Create snapshot",
	Long: `Sends CSI create snapshot requirment

Regular CSI create snapshot request.`,
	Run: createSnaposhot,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	createSnapshotCmd.Flags().StringVarP(&snapshotName,"csn",	"", "", "snapshot name")
	createSnapshotCmd.Flags().StringVarP(&volumeId,"sv",	"", "", "volume id")

	if err:= createVolumeCmd.MarkFlagRequired("sv"); err != nil {
		fmt.Println(err)
	}
	
	if err:= createVolumeCmd.MarkFlagRequired("csn"); err != nil {
		fmt.Println(err)
	}
	
	//if err:= createVolumeCmd.MarkFlagRequired("size"); err != nil {
	//	fmt.Println(err)
	//}

	ControllerCmd.AddCommand(createSnapshotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
