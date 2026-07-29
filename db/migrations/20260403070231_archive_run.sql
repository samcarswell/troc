-- migrate:up
alter table runs
add column is_archived boolean not null default false;

-- migrate:down
alter table crons
drop column is_archived;
