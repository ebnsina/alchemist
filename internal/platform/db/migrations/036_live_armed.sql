-- Arming is not broadcasting.
--
-- A stream was armed by creating its asset in 'live', so a class nobody had started
-- publishing to showed as ON AIR in the video list and handed out playback URLs for
-- segments that did not exist yet. MarkLive already runs on the first segment; this
-- is the state the asset sits in until then.
alter type asset_state add value if not exists 'live_armed' before 'live';
