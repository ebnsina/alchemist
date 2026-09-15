-- A customer publishing from their own browser, with no encoder at all.
--
-- The column names what the customer chose, not what goes over the wire: 'camera' is
-- "my webcam", and srt/rtmp are the two answers to "my encoder". A teacher with a
-- laptop and no OBS could not use the product at all before this.
--
-- Additive on purpose. The browser publishes WebRTC over WHIP to the same path the
-- ingest server already authorises by (key, path), so the transcoder's loopback RTSP
-- pull, the origin, the signing and the recording path are all unchanged.
alter table live_streams drop constraint live_streams_protocol_check;
alter table live_streams add constraint live_streams_protocol_check
  check (protocol in ('srt', 'rtmp', 'camera'));
