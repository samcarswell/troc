# Troc

Simple cron (or any script) monitoring

<img src="static/the-crons-friend.jpg" alt="The Cron's Friend" width="512"/>

*The Cron's Friend* - Named after the legendary bird [Trochilus](https://en.wikipedia.org/wiki/Trochilus_(crocodile_bird)) described by Herodotus in [The Histories](https://en.wikipedia.org/wiki/Histories_(Herodotus)).

## Features

`troc` is a wrapper for shell commands that provides the following:
- Automatically stores stdout/stderr logs of jobs.
- Watch and tail stdout/stderr of running or past jobs.
- Keeps a history of all job runs in a local sqlite database.
- Query job runs using the `troc` cli.
- Ensures that only one instance of a job is ran at a time; keeps a log of skipped runs.
- Notifications. `ntfy` and `slack` support is built-in, but custom systems can be configured.
- Single executable; no daemon.

## Build

Checkout the relevant release tag. eg. `0.5.0`; you can build on any commit but there might be issues.

```bash
git checkout 0.5.0
./build
```

Will create a `troc` binary in root directory of the repo. You can output elsewhere if you like: `./build ~/.local/bin`.

## Installation

Ensure the `troc` binary is available in your path. eg. `PATH=$PATH:~/.local/bin` assuming `troc` is in `~/.local/bin`.

Executing `troc` for the first time will setup your config and database with the default settings:

```log
fsh ❯ troc exec --name "test" "echo 'Testing...'"
time=2025-11-09T14:02:31.182+11:00 level=INFO msg="Creating config directory at /home/srcarswell/.config/troc"
time=2025-11-09T14:02:31.182+11:00 level=INFO msg="Creating initial config file at /home/srcarswell/.config/troc/config.yaml"
time=2025-11-09T14:02:31.184+11:00 level=INFO msg="Logging to /tmp/trocsys_wfnzz_20251109T030231.log"
time=2025-11-09T14:02:31.184+11:00 level=INFO msg="Creating: /home/srcarswell/.config/troc/troc.db"
time=2025-11-09T14:02:31.187+11:00 level=INFO msg="Applying: 20251104051400_initial.sql"
time=2025-11-09T14:02:31.189+11:00 level=INFO msg="Applied: 20251104051400_initial.sql in 1.886512ms"
time=2025-11-09T14:02:31.189+11:00 level=INFO msg="Applying: 20251108020521_message_cron.sql"
time=2025-11-09T14:02:31.191+11:00 level=INFO msg="Applied: 20251108020521_message_cron.sql in 2.176757ms"
time=2025-11-09T14:02:31.195+11:00 level=INFO msg="Job not registered. Creating new Job with name test"
time=2025-11-09T14:02:31.196+11:00 level=INFO msg="Run log created at: /tmp/test.4227158531.log"
time=2025-11-09T14:02:31.197+11:00 level=INFO msg="Run created with ID 1"
time=2025-11-09T14:02:31.199+11:00 level=INFO msg="Run 1 completed: Succeeded"
{
    "ID": 1,
    "JobName": "test",
    "StartTime": "2025-11-22 12:46:01 +1100 AEDT",
    "EndTime": "2025-11-22 12:46:02 +1100 AEDT",
    "LogFile": "/tmp/test.4227158531.log",
    "SystemLogFile": "/tmp/trocsys_wfnzz_20251109T030231.log",
    "Status": "Succeeded"
    "Duration": "1s"
}
```

Subsequent runs won't do this:

```log
time=2025-11-09T14:03:18.256+11:00 level=INFO msg="Logging to /tmp/trocsys_yncwi_20251109T030318.log"
time=2025-11-09T14:03:18.261+11:00 level=INFO msg="Run log created at: /tmp/test.2330262208.log"
time=2025-11-09T14:03:18.262+11:00 level=INFO msg="Run created with ID 2"
time=2025-11-09T14:03:18.264+11:00 level=INFO msg="Run 2 completed: Succeeded"
{
    "ID": 2,
...
```

Updating `troc` versions may also need to apply migrations to your database,
in which case these will be logged similarly to the first run. eg.

```log
...
time=2025-11-09T14:03:18.240+11:00 level=INFO msg="Applying: 20251108020521_a_new_migration.sql"
...
```

### Config

Running `troc` for the first time will create a config file at `~/.config/troc/config.yaml`
if it does not exist with default values.

Any of the config values can also be specified using env vars:
eg. `TROC_DATABASE` or `TROC_NOTIFY_SLACK_TOKEN`.


| Name | Description | Default |
| - | - | - |
| `database` | Path to the sqlite database. | `~/.config/troc/troc.db` |
| `localtime` | Display dates in local time rather than UTC. | `true` |
| `logdir` | Directory of job log files. | `$TMPDIR` if not empty, otherwise `/tmp` |
| `logjson` | Output stderr system logs in json format. Note: if defined in `$HOME/.config/troc/config.yaml` this will only take affect after configuration has been loaded. Any logging that occurs before this, such as startup failures, will be in text format. If you are running `troc` in an automated fashion and are relying on stderr system logs being in a json format, ensure that the env var `TROC_LOGJSON=true` is set; this will affect log format immediately. | `false` |
| `notify.hostname` | Name of server when pushing notifications. eg. `job-name@hostname` | Output of `hostname` |
| `notify.system` | Determines the notification system to use. | `ntfy` |
| `notify.ntfy.topic` | Ntfy topic to post notifications. | |
| `notify.ntfy.domain` | Ntfy domain. | `https://ntfy.sh` |
| `notify.ntfy.token` | Ntfy token. Not required for default domain. |  |
| `notify.slack.token` | Token for slack app. | 
| `notify.slack.channel` | Slack channel to post notifications. | 
| `notify.status.succeeded` | Sets message as a priority: either `@channel` for Slack or `Priority=3` for ntfy, for `Succeeded` status. | `false` |
| `notify.status.failed` | Sets message as a priority: either `@channel` for Slack or `Priority=3` for ntfy, for `Failed` status. | `true` |
| `notify.status.running` | Sets message as a priority: either `@channel` for Slack or `Priority=3` for ntfy, for `Running` status. | `false` |
| `notify.status.skipped` | Sets message as a priority: either `@channel` for Slack or `Priority=3` for ntfy, for `Skipped` status. | `false` |
| `notify.status.terminated` | Sets message as a priority: either `@channel` for Slack or `Priority=3` for ntfy, for `Terminated` status. | `true` |
| `display.emoji` | Displays emojis. | `true` |
| `display.color.status.succeeded` | Colours text output for `Succeeded` status. | `false` |
| `display.color.status.failed` | Colours text output for `Failed` status. | `false` |
| `display.color.status.running` | Colours text output for `Running` status. | `false` |
| `display.color.status.skipped` | Colours text output for `Skipped` status. | `false` |
| `display.color.status.terminated` | Colours text output for `Terminated` status. | `false` |
| `clean.days` | Number of days to keep runs when running `troc clean` | `30` |

Any invocation of `troc` will check for a database located at the `database` config value.
If it does not exist, it will create it.

## Usage

### Running a job

`troc exec` handles the execution of a job. Use `--name` to specify the name of
the job. If it does not exist, it will be created with the default settings;
for anything non-default use `troc job add` before `troc exec`.

Use the args to pass the command you intend to run.

eg. `troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir"`

This will create (if it does not exist) a job named `daily-sync` and will
execute `rsync --avh /tmp/source-dir /tmp/dest-dir` as a run of that job.

The stdout log will display the id of the run:
```
...
time=2025-11-04T18:03:44.964+11:00 level=INFO msg="Run created with ID 1"
...
```

You can use this id to see the ongoing (for long-running jobs) or completed logs:
`troc run show -r 1`
Output:
```json
{
    "ID": 1,
    "JobName": "daily-sync",
    "StartTime": "2025-11-22 12:46:01 +1100 AEDT",
    "EndTime": "2025-11-22 12:46:02 +1100 AEDT",
    "LogFile": "/tmp/daily-sync.3159256558.log",
    "SystemLogFile": "/tmp/trocsys_pgqlq_20251104T070344.log",
    "Status": "Succeeded",
    "Duration": "1s"
}
```

If you have the `notify.slack.*` config values, you can append `--notify` to the `troc exec`
command to send a notification in slack:

```markdown
*daily-sync@example-server:84* - ✅: Succeeded
Log: /tmp/daily-sync.3159256558.log
```

### Watching a run

Use `troc run watch -r [RUN_ID]` to tail the logs of a running job until it completes. If the job has already ran, it will print the logs and immediately exit.

### Details of a run

`troc run show -r [RUN_ID]`

### Kill a run

To kill a run, use `troc run kill -r [RUN_ID]`. This will print the PID of
the run, and an interactive prompt to confirm killing the PID. 

If you are using the command in an automated way, you can provide the `--force`
flag.

`troc run kill` will pass the `SIGTERM` down to the executing script, 
and handle terminating the run by setting it's state to `Terminated`.

#### SIGKILL

If a `troc exec` process was killed using `SIGKILL`, eg. `kill -9 [PID]`,
it cannot be gracefully handled. The script passed to `troc exec` will be
left running and the run itself will be left in a `Running` state.

To cleanup the run, you can manually set its state to `Terminated` 
using `troc run term -r [RUN_ID]`.

Note that `troc run term` will not check if the process is still running, 
or attempt to terminate it itself.
Only run this if you have determined that the run is not running and 
it's state is still `Running`.

If the run is still in progress, and `troc run term` has been ran on it,
the run will still correctly update it's state once it completes.

### Archive a run

Use `troc run archive` to archive a run; this will remove all log files,
and hide it from appearing in `troc run list`.

### Run history

Use `troc run list` to see a list of historical runs. Optionally filter on `--name`,
or `--include-archived`.

### Update job info

A job name and log settings can be updated using `troc job update`.

### Outputting run output to slack

Configure the relevant job using `--notify-log` with the `troc job [add|update]` command.
This will include the raw run output in the slack message:

```markdown
daily-sync@example-server: run 84 - ✅: Succeeded
This is output from the job
```

### Clean old runs and system log files

Use `troc clean` to archive all runs older than `clean.days`, and remove
any miscellaneous system log files.

### Crontab example:

```
PATH=$PATH:/usr/local/bin:/usr/bin # Ensuring that troc and rsync is in the path

*/5 * * * * troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir" --notify
```

### Run states

| Name | Description |
| - | - |
| `Running` | Run is still actively running. |
| `Skipped` | The run was skipped as there is already another run of the same job in progress. |
| `Succeeded` | The run completed with an exit code == 0. |
| `Failed` | The run completed with an exit code != 0. |
| `Terminated` | The run received a `SIGINT` or `SIGTERM`. |


### Notifications

Runs can optionally send notification messages using the `--notify` flag on `troc exec` to third-party systems.
By default `troc` can send messages to `slack `or `ntfy`, but any system can be used provided it's accessible programmatically. 
The priority of these notifications is determined by config values `notify.status.*`.
If a status is set to `true`, this status will cause:
- A priority of 3 in `ntfy`.
- Tagging of the channel in `slack`.

#### Custom systems

Custom notification systems can be set using config values `notify.custom.[INDEX].[name|command|envvars]` where:
- `name` sets the unique identifier for this notification system.
- `command` sets the command to be run on notify.
- `envvars` sets environment variables to be available in the command.

`troc` will pass the following json object as stdin to `command`:

```json
{
  "run": {
    "id": 23902,
    "job_name": "say-hello",
    "start_time": "2026-08-17 15:24:04 +0930 AEST",
    "end_time": "2026-08-17 15:24:04 +0930 AEST",
    "log_file": "/tmp/log-output.1308558633.log",
    "system_log_file": "/tmp/trocsys_mtnve_20260817T055404.log",
    "status": "Failed",
    "duration": "0s",
    "pid": "82459",
    "is_archived": false
  },
  "job": {
    "id": 63,
    "name": "say-hello",
    "notify_log_content": true
  },
  "status_config": {
    "succeeded": false,
    "failed": true,
    "running": false,
    "skipped": false,
    "terminated": true
  },
  "hostname": "fedora"
}
```

Where:
- `run` is the run information.
- `job` is the job information.
- `status_config` is the `notify.status` configuration value.
- `hostname` is the `notify.hostname` configuration value.

##### Example using `wall`

`~/.config/troc/config.yaml`:
```yaml
...
notify:
    system: wall # Using our custom 'wall' notify
    custom:
      - name: wall
        command: /tmp/wall.sh
        envvars:
          - WALL_MSG=A job has been run
...
```

`/tmp/wall.sh`:
```bash
#!/bin/bash

# Read stdin json object
readarray text 

# Read values
JOB=$(echo $text | jq -r '.job.name')
ID=$(echo $text | jq -r '.run.id')
STATUS=$(echo $text | jq -r '.run.status')
MESSAGE="$WALL_MSG: $ID:$JOB - $STATUS"

# Send to third-party system 
wall "$MESSAGE"

```

Executing `troc exec --name say-hello --notify "echo 'Hello!'"` will cause the following popup on all shells:

```
Broadcast message from user@fedora (somewhere) (Mon Aug 17 15:26:49 2026)

A job has been run: 23905:say-hello - Succeeded
```

## Troubleshooting

If `exec` fails to create a run, it errored before it could create the run.
This is probably caused by configuration issues. You can debug this by looking through the
system logs: `[logdir]/trocsys*.log`

## Development

### SQL queries

All query code is generated from `./db/query.sql` using [sqlc](https://github.com/sqlc-dev/sqlc).

To update/add a query, make the changes needed to `./db/query.sql` and then run
`sqlc generate`. This will generate code in `./data`; no non-generated code
should be placed in this directory.

### Database migrations

`./migration [migration_name]`

This will create a new migration file in `./db/migrations/` which can be updated
to include the raw sql migration statements.

All `./db/migrations/*sql` files are embedded into the binary and ran
at startup; so just adding the migration file to that directory is enough.

### Running tests

`go test ./...`

## FAQ

### Is this a replacement for cron?

No, `troc` just wraps a regular shell command. It's best paired with cron
but you can run it independently.

### Will this run on Windows?

Not sure. Feel free to try.

### How can I communicate run information to another system?

See [Custom Systems](#custom-systems) to execute arbitrary scripts after a run.

## Future goals

- TUI interface.
- A local `troc` should be able to connect to a remote `troc` using SSH.
