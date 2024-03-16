/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package node

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Comand line interface to running node commands",
	Long: `A long description of node command goes here
and continue here.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("node called")
	},
}



func init() {
	//rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
