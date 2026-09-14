-- Mark the expensive rungs lazy. The low ladder is always present so playback works
-- instantly for every asset; 480p and above are generated on first request.
--
-- Sized from the BD case: 144-360p costs ~0.9 GB per source hour, the full ladder
-- ~5.2 GB. Most of a library is never watched, so most of that is waste.
update ladder_profiles set rungs = (
  select jsonb_agg(
    case when (r->>'height')::int >= 480
         then r || '{"lazy":true}'::jsonb
         else r end
    order by (r->>'height')::int)
  from jsonb_array_elements(rungs) r
) where name in ('bd-mobile', 'bd-ott', 'intl-default');

-- Renditions that have not been generated yet are recorded up front so the API can
-- report what exists and what is pending without re-deriving it from the profile.
alter table renditions add column if not exists lazy boolean not null default false;

-- When playback first asked for a lazy rendition. Null means nobody has.
alter table renditions add column if not exists requested_at timestamptz;
