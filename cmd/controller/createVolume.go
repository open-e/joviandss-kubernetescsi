/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package controller

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	//humanize "github.com/dustin/go-humanize"

	csi_common "joviandss-kubernetescsi/pkg/common"
	csi_controller "joviandss-kubernetescsi/pkg/controller"
	cli_common "joviandss-kubernetescsi/pkg/common"
	
	"joviandss-kubernetescsi/pkg/common"

	"github.com/spf13/cobra"
)

var (
	//volumeName string
	//volumeSize string
	
	//volumeSizeRequired string
	//volumeSizeLimit string
)
	

func createVolume(cmd *cobra.Command, args []string) {
	logrus.Debug("create volume")
	
	var cfg csi_common.JovianDSSCfg

	var cp csi_controller.ControllerPlugin

	if err := csi_common.SetupConfig(cli_common.ControllerConfigPath, &cfg) ; err != nil {
		// GetConfig(ControllerConfigPath, &controllerCfg)
		panic(err)
	}
	csi_controller.SetupControllerPlugin(&cp, &cfg)

	// var vol csi_rest.Volume = csi_rest.Volume{Name: "test-1", Size: "1G"}

	var req csi.CreateVolumeRequest
	var ctx context.Context = common.GetContext("create_volume")
	
	var bytes uint64 = 0


	var cr csi.CapacityRange
	cr.RequiredBytes = int64(bytes)
	req.CapacityRange = &cr

	req.Name = volumeName
	var supportedVolumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
		// VolumeCapability_AccessMode_UNKNOWN,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		// VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		// VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		// VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}

	req.VolumeCapabilities = csi_controller.GetVolumeCapability(supportedVolumeCapabilities)
        
	if len(sourceSnapshotName) > 0 {
		req.VolumeContentSource = &csi.VolumeContentSource{
        	    Type: &csi.VolumeContentSource_Snapshot{
        	        Snapshot: &csi.VolumeContentSource_SnapshotSource{
        	            SnapshotId: sourceSnapshotName,
        	        },
        	    },
        	}
	}

	if len( sourceVolumeName) > 0 {
		req.VolumeContentSource = &csi.VolumeContentSource{
        	    Type: &csi.VolumeContentSource_Volume{
        	        Volume: &csi.VolumeContentSource_VolumeSource{
        	            VolumeId: sourceVolumeName,
        	        },
        	    },
        	}
	}

	if volumeSizeLimit != 0 {
		req.CapacityRange = &csi.CapacityRange{
			LimitBytes: volumeSizeLimit,
		}
		
	}

	if volumeSizeRequired != 0 {
		req.CapacityRange = &csi.CapacityRange{
			RequiredBytes: int64(volumeSizeRequired),
		}
	}
	resp, err := cp.CreateVolume(ctx, &req)
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
var createVolumeCmd = &cobra.Command{
	Use:   "createVolume",
	Short: "Create volume",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: createVolume,
	//func(cmd *cobra.Command, args []string) {
	//	fmt.Println("createvolume called")
	//},
}

func init() {

	createVolumeCmd.Flags().StringVarP(&volumeName,"name",	"n",		"", "Name of volume to create")
	//createVolumeCmd.Flags().StringVarP(&volumeSize,"size",	"s",		"", "Size of volume to create")
	createVolumeCmd.Flags().Int64VarP(&volumeSizeRequired,	"srq",		"", 0, "Required size of volume to create")
	createVolumeCmd.Flags().Int64VarP(&volumeSizeLimit,	"slm",		"", 0, "Limit size of volume to create")
	createVolumeCmd.Flags().StringVarP(&sourceVolumeName,	"volume",	"", "", "Name of source volume to use")
	createVolumeCmd.Flags().StringVarP(&sourceSnapshotName,	"snapshot",	"", "", "Name of source snapshot to use")

	if err:= createVolumeCmd.MarkFlagRequired("name"); err != nil {
		fmt.Println(err)
	}
	
	//if err:= createVolumeCmd.MarkFlagRequired("size"); err != nil {
	//	fmt.Println(err)
	//}

	ControllerCmd.AddCommand(createVolumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createvolumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createvolumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
