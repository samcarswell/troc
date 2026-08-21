# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2026-08-21

### Added

- Notifications can now be sent by `ntfy`.
- Notifications can now be sent by custom scripts.
- `job add` prints job info to stdout.

### Changed

- Notifications default to using `ntfy`. Set `notify.system` to `slack` to use Slack.

### Removed

- Config value `lockdir`. Database querying for running jobs has replaced lock files.

### Fixed

- Runs that fail to notify correctly print run information to stdout.

## [0.5.0] - 2026-07-29

### Added

- `run archive` to archive a single run.
- `clean` to archive old runs and remove system log files.

### Changed

- Notification output text has been simplified.

### Fixed

- Issue where watching a run with a removed log file would return a zero exit code.

## [0.4.1] - 2026-06-16

### Changed

- More detailed failure logging for database connection creation.

### Fixed

- Issue where `SQLITE_BUSY_RECOVERY` could occur during database connection creation.
- Slack errors are now included in run logs.

## [0.4.0] - 2026-04-01

### Added

- Configure which statuses tag `@channel` on notify using config value `notify.status.[succeeded|failed|running|skipped|terminated]`.

### Fixed

- Issues with many simultaneous runs.

## [0.3.0] - 2025-11-30

### Added

- Displaying of emojis using config value `display.emoji`. Defaults to `true`.
- Colour run status output using config value `display.color.status.[succeeded|failed|running|skipped|terminated]`. Defaults to `false`.
- Handling SIGTERM in `exec` by setting run state to `Terminated`.
- `run kill` to kill runs with status `Running`.
- `--format` options for `[run|job] list`.
- Config value `logjson=true` will output stderr logs in json. Defaults to `false`.

## [0.2.0] - 2025-11-23

### Added

- Handling SIGINT in `exec` by setting run state to `Terminated`.
- Filtering by job in `run list`.
- `job update` command.
- `run term` command to manually set run state to `Terminated` for orphaned `Running` runs. eg. runs that have been killed using SIGKILL, since we cannot gracefully handle it.

### Changed

- Moved `cron [add|list|update]` to `job [add|list|update]`.
- Display times in local timezone. This can be made UTC again using config value: `localtime=false`.

### Fixed

- Issues using `exec` with commands that chain or pipe multiple commands.

### Removed

- Jobs configured with `--notify-log` only notify the raw log content; the log path has been removed.

## [0.1.1] - 2025-11-09

### Fixed

- Issue where database migration logs were sent stdout, breaking JSON output.
- Uniform formatting for logs. Stderr is formatted as text, files are JSON.
- `run watch` now fails for non-existent runs.

## [0.1.0] - 2025-11-09

### Added

- Allow crons to be configured to include log contents in notify messages.
- `lockdir` can now be used to configure lock directory.
- `--version` option to display `troc` version.

### Changed

- Moved `hostname` config to `notify.hostname`.
- Moved `slack.channel` config to `notify.slack.channel`.
- Moved `slack.token` config to `notify.slack.token`.

### Fixed

- Some concurrency issues with the database. 

## [0.0.1] - 2025-11-04

### Added

- `exec` command to run crons.
- `run list` command to list cron runs.
- `run watch` command to tail logs of runs.
- `cron add` command to add crons.
- `cron list` command to list crons.
- `run show` command to display run.

