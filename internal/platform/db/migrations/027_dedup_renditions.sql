-- A deduplicated asset owned rendition rows describing media it does not own.
--
-- The copies were a lie with a cost: inheriting the canonical asset's pending lazy
-- rows meant first play of the duplicate queued a JIT encode for the duplicate, which
-- wrote cmaf/{tenant}/{duplicate}/ -- a prefix no playback request ever resolves to.
-- The rung burned CPU, was served to nobody, and the canonical asset's own rung stayed
-- pending forever. Renditions are now read through deduplicated_from, so the copies
-- have to go or they shadow nothing and confuse everything.
delete from renditions r
 using assets a
 where a.id = r.asset_id and a.deduplicated_from is not null;
