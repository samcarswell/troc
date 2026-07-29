/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"github.com/samcarswell/troc/cmd"
	"github.com/spf13/cobra"
)

var JobCmd = &cobra.Command{
	Use:   "job",
	Short: "Commands related to jobs",
}

func init() {
	cmd.RootCmd.AddCommand(JobCmd)
}
