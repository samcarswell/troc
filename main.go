/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	"embed"

	"github.com/samcarswell/troc/cmd"
	_ "github.com/samcarswell/troc/cmd/clean"
	_ "github.com/samcarswell/troc/cmd/exec"
	_ "github.com/samcarswell/troc/cmd/job"
	_ "github.com/samcarswell/troc/cmd/run"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	cmd.Execute(migrations)
}
