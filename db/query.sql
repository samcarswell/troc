-- name: GetJob :one
select
    sqlc.embed(jobs)
from jobs
where jobs.name = ?;

-- name: CreateJob :one
insert into jobs
    (name, notify_log_content)
values (?, ?)
returning id;

-- name: StartRun :one
insert into runs
    (job_id, start_time, log_file, exec_log_file, status)
values (?, current_timestamp, ?, ?, "Running")
returning id;

-- name: EndRun :exec
update runs
set end_time = current_timestamp, status = ?
where id = ?;

-- name: SkipRun :one
insert into runs
    (job_id, start_time, end_time, log_file, exec_log_file, status)
values (?, current_timestamp, current_timestamp, "", ?, "Skipped")
returning id;

-- name: GetJobs :many
select
    sqlc.embed(jobs)
from jobs;

-- name: GetRuns :many
select
    sqlc.embed(runs),
    sqlc.embed(jobs)
from runs, jobs
where runs.job_id = jobs.id
and (?1 = '' or jobs.name = ?1)
and (?2 or runs.is_archived = ?2);

-- name: GetRun :one
select
    sqlc.embed(runs),
    sqlc.embed(jobs)
from runs, jobs
where runs.job_id = jobs.id
and runs.id = ?;

-- name: IsRunFinished :one
select runs.end_time is not null
from runs
where runs.id = ?;

-- name: UpdateJob :exec
update jobs
set name = ?2, notify_log_content = ?3
where id == ?1;

-- name: UpdateRunPid :exec
update runs
set pid = ?2
where id == ?1;

-- name: GetNonArchivedRunsBeforeDate :many
select
    sqlc.embed(runs),
    sqlc.embed(jobs)
from runs, jobs
where runs.job_id = jobs.id
and is_archived = false
and runs.end_time < ?1;

-- name: ArchiveRun :exec
update runs
set is_archived = true
where id = ?;
