# herdr-mobile and original skid still share installation identity

problem: the herdr-backed product and preserved original tmux product claim
the same github release destination, android package, host paths, service
names, ports, credentials, and deployment ownership. the planned github name
reclamation also removes the old-name redirect.

impact: installing or maintaining either can replace or disrupt the other.
separate provider homes alone do not prevent project-level hook discovery;
the original deployment also lacks its retired native helper.

evidence (2026-09-25): source inspection at herdr-backed `d8bb9c4`, original
`927c554`, dev-server `8498933`, jarvis `c739f2d`; read-only observations of
all three hosts and the usb phone are recorded in the
[separation spec](../herdr-mobile-separation.md#2-investigated-baseline).
the phone holds herdr-backed `0.8.0` under `dev.niels.skidbladnir`.

reproduction: compare both repositories' command defaults, android application
id and release destination with dev-server's installed service/path/port
ownership. do not reproduce by installing one over the other.

source progress: this repo now uses its separate command/module, package,
signer, paths, ports, headers, invitation kind, metadata, and release tooling.
engineering checks, signed candidate build, isolated native gateway/metadata
probe, and compiled android origin/invitation probe passed; details are in the
[deployment contract](../dev-server-handoff.md). the obsolete release pin was
removed; no replacement pin is invented before publication.

remaining work: finish the cross-repo spec, original-repo handoff, and
[dev-server handoff](../dev-server-separation-handoff.md); qualify private
provider homes/authentication/hook trust and native-helper compatibility;
verify new-port reachability from the phone; inventory publishing clones and
obsolete host generations; coordinate jarvis workers, github names, namespace
handback, phone reset/re-pairing, and separated rollback/removal.
github names were transferred successfully on 2026-09-25; both repository
ids were preserved and known clone remotes on all three hosts retargeted.

known blockers to claiming completion: no separated release, deployment,
provider setup, or fleet/phone coexistence proof exists yet. credential/login
requirements and current helper compatibility are unproven. fleet and phone
acceptance remains `NOT_RUN`; isolated probes do not prove deployed behavior.

resolved when: every acceptance row in the spec passes for both products,
on all applicable hosts and the phone, and neither ordinary installation nor
maintenance depends on the other's owned state. delete this issue then.
