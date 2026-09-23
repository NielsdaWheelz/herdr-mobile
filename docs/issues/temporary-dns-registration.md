# temporary phone-fixture dns registration

problem: the pr 134 physical-phone fixture registered `bright-garden.lancert.dev`
and its wildcard to the mac's private lan address. the phone journey was stopped
before any pairing invitation was created or scanned. test servers, proxy, proof
app and local certificate/key were removed, but lancert v2 exposes no public
operation to delete or disable the registered address records.

impact: the private lan address remains visible in public dns. it is not a
publicly routable service. the issued certificate also has a permanent public
certificate-transparency entry; deleting local files cannot remove that entry.

evidence: [lancert's published security model](https://github.com/lucor/lancert#security-model)
limits registration credentials to short-lived acme txt values and says they
cannot change address records. its public api has registration and txt update,
without a deletion operation. the registered address has no automatic expiry.
the qr display, phone proxy and adb tunnel were removed and the five prior
unset proxy settings were verified absent. no invite was created or scanned.

resolution owner: the lancert operator. resolved when the operator removes or
disables this registration and authoritative dns no longer answers with the
mac's private lan address. the certificate-transparency entry is irreversible.
the user declined external contact, so no removal request was sent.
