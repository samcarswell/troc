package opts

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/samcarswell/troc/core"
	"github.com/spf13/cobra"
)

func GetStringOptOrExit(cmd *cobra.Command, name string) string {
	optVal, err := cmd.Flags().GetString(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	return optVal
}
func GetBoolOptOrExit(cmd *cobra.Command, name string) bool {
	optVal, err := cmd.Flags().GetBool(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	return optVal
}
func GetInt64OrExit(cmd *cobra.Command, name string) int64 {
	optVal, err := cmd.Flags().GetInt64(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	if optVal == 0 {
		core.LogErrorAndExit(slog.Default(), errors.New("option "+name+" must have a value"))
	}
	return optVal
}

func FormatTableOpt(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "format", "f", string(core.FormatPretty), "Format output (pretty|json|csv|tsv)")
}

func FormatObjectOpt(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "format", "f", string(core.FormatJson), "Format output (json)")
}

func FormatTableOptValidate(cmd *cobra.Command, format string) error {
	switch core.OutputFormat(format) {
	case core.FormatPretty, core.FormatCsv, core.FormatTsv, core.FormatJson:
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
	return nil
}

func FormatObjectOptValidate(cmd *cobra.Command, format string) error {
	switch core.OutputFormat(format) {
	case core.FormatJson:
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
	return nil
}
