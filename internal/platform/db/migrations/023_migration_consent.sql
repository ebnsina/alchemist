-- Nothing is imported until the customer has seen the list and said yes. A migration
-- now starts in 'previewing': items are recorded, no assets and no transcodes.
alter table migration_sources drop constraint migration_sources_state_check;
alter table migration_sources add constraint migration_sources_state_check
  check (state in ('previewing', 'scanning', 'done', 'failed', 'paused'));
alter table migration_sources alter column state set default 'previewing';

-- 'previewing' alone cannot say whether the list is still being fetched or is
-- complete and waiting on a person, and those are different screens.
alter table migration_sources add column preview_done boolean not null default false;
