-- migrate:up
CREATE TRIGGER one_running_row
BEFORE INSERT ON runs
FOR EACH ROW
WHEN (
    EXISTS (SELECT NULL
            FROM runs
            WHERE status = "Running"
            AND job_id = NEW.job_id
           )
    and NEW.status != "Skipped"
)
BEGIN
  SELECT RAISE( ABORT, 'already a run for this job with status Running' );
END;
-- migrate:down
drop trigger one_running_row;
