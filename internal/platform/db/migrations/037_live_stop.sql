-- Ending a broadcast from the dashboard.
--
-- Until now a stream stopped only when the encoder disconnected or the reaper timed
-- it out, so a customer who closed OBS badly, or wanted to end a class early, had
-- nothing. The API cannot reach into the worker's ffmpeg, so it writes the intent
-- here and the worker's own poll picks it up on the next tick.
alter table live_sessions add column if not exists stop_requested_at timestamptz;
