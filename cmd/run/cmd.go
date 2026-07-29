/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"github.com/samcarswell/troc/cmd"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Commands related to runs",
}

func init() {
	cmd.RootCmd.AddCommand(RunCmd)
}
