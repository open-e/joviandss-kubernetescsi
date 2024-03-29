/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package node

import (
	"fmt"
	//"os"

	"github.com/spf13/cobra"
	//cli_common "joviandss-kubernetescsi/pkg/common"
	
	cliGetInfo "joviandss-kubernetescsi/cmd/node/getinfo"
	cliStageVolume "joviandss-kubernetescsi/cmd/node/stagevolume"

)

// nodeCmd represents the node command
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Comand line interface to node commands",
	Long: `That is general sub command that stores all node related commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("node called")
	},
}

func addSubCmds() {
	NodeCmd.AddCommand(cliGetInfo.NodeGetInfoCmd)
	NodeCmd.AddCommand(cliStageVolume.NodeStageVolumeCmd)
}


func init() {
	
	addSubCmds()
	//rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
