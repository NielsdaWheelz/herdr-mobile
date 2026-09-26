# phone tailscale dns stops answering

2026-09-25: the phone ran tailscale android `1.102.3`. direct tailnet-ip tcp
connections to ports `8443` and `8444` worked on all three hosts, while their
hostnames timed out. icmp to `100.100.100.100` worked. a raw udp probe was
inconclusive even after recovery and is excluded from the evidence.
pairing therefore stopped at a generic fleet connection
error before either app could be qualified on the phone.

force-stopping and relaunching the tailscale process, then reconnecting,
restored hostname tcp connections to all three hosts immediately. this is a
process recovery, not evidence that the fault cannot recur after a network
change. the pattern likely matches [tailscale issue #21155](https://github.com/tailscale/tailscale/issues/21155),
but the phone's tun error and peerapi behavior were not observed, so the cause
is unproven.

[the fix](https://github.com/tailscale/tailscale/pull/21332) merged on september 21
and was [backported to `release-branch/1.102`](https://github.com/tailscale/tailscale/pull/21445)
on september 24. [current branch source](https://github.com/tailscale/tailscale/blob/release-branch/1.102/wgengine/netstack/netstack.go#L1073-L1109)
retains it. the [latest published stable android release](https://github.com/tailscale/tailscale-android/releases)
and [official stable apk](https://dl.tailscale.com/stable/) are still `1.102.4`,
which predates the fix. no fixed stable android upgrade is confirmed available.

resolution: confirm a published stable android build contains the fix, update
through the [official play store route](https://tailscale.com/docs/install/android),
then check quad-100 dns replies and hostname tcp to all three hosts across a
wifi/cellular transition. retry both apps' pairing and phone journeys after
the network boundary holds.
