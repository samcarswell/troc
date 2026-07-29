package lib

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/samcarswell/troc/data"
)

func ArchiveRun(
	ctx context.Context,
	queries *data.Queries,
	run data.Run,
	logger *slog.Logger,
) error {

	runIdStr := strconv.Itoa(int(run.ID))
	logger.Info("Archiving run " + runIdStr)
	err := queries.ArchiveRun(ctx, run.ID)
	if err != nil {
		return fmt.Errorf("failed to archive run: %w", err)
	}
	logErr := os.Remove(run.LogFile)
	if logErr != nil {
		logger.Error("Failed to delete run log file", "err", logErr)
	}
	sysLogErr := os.Remove(run.ExecLogFile)
	if sysLogErr != nil {
		logger.Error("Failed to delete run system log file", "err", sysLogErr)
	}
	if logErr != nil || sysLogErr != nil {
		logger.Warn("Run " + runIdStr + " was archived, but failed to remove some logs")
	} else {
		logger.Info("Run " + runIdStr + " successfully archived.")
	}
	return nil
}
