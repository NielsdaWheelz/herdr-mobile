# herdr-mobile launcher mark

2026-09-25. the herdr phone package uses a gate-shaped `h` on the existing
niðavellir ink background. it identifies the herdr-backed product beside the
original skíðblaðnir app. the dwarf catalogue, seals, type, and screen palette
remain unchanged.

`scripts/gen-ornament` owns the adaptive icon resources. the 108-unit canvas
has two ten-unit gold uprights joined by a ten-unit bridge. each outer end has
a five-unit cut corner, echoing the app's faceted cards. the foreground colours
the bridge bronze. the monochrome drawable uses the same solid silhouette,
which remains legible when android applies a single tint. the background keeps
the ink and deep-surface diagonal facet. the farthest foreground vertex is
inside the 34-unit radius that precedes android's circular mask at 36 units.

at the minimum 48dp launcher size, the narrowest stem and bridge each render
at 6.67dp. a glance should read the `h` as one built gate, without the ship's
sail or hull silhouette. the mark is geometric, not a second dwarf glyph.

this replaces the skíðblaðnir ship mark introduced by commit `92cb6ad`.
that mark remains part of the original product's design history; no released
artifact is changed. generation is deterministic, and
`scripts/gen-ornament --check` verifies the checked-in output.
