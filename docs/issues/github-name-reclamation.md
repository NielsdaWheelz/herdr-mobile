# github name reclamation is not yet proven

problem: the plan renames repository id `1342599607` away from
`NielsdaWheelz/skidbladnir`, then gives that name to original repository id
`1386409483`. github can permanently retire old repository namespaces.

impact: a successful first rename does not by itself establish that the
original repository can reclaim the requested name.

evidence (2026-09-25): the current repo's traffic api reports 1,211 clones,
including 685 in the last seven reported daily buckets.
[github's explanation](https://github.blog/security/supply-chain-security/how-to-stay-safe-from-repo-jacking/)
describes usage-based retirement. detailed docs specify transfer/account-rename
thresholds; applicability to this same-owner repository-name swap remains
uncertain. no rename/reuse attempt was made, and no rejection is claimed.

follow-up: establish policy applicability before depending on reclamation.
the root operator owns the eventual name change and repository-id checks.
if refused, retain both repositories/histories and original `skid-v1` while
resolving the naming requirement; do not delete/recreate or replace history.

resolved when: github accepts `herdr-mobile` for id `1342599607` and
`skidbladnir` for id `1386409483`, with consumers retargeted as specified in
the [cutover plan](../herdr-mobile-separation.md#b-transfer-github-names).
