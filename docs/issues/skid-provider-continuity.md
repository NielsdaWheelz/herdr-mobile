# original skid's account correction awaits installer convergence

problem: original skid `3bd0ae4` and the corrected separation spec now select
existing accounts, but dev-server `1296309` still provisions private skid
homes. those revisions do not yet form a deployable pair.

impact: skid launches, including bare provider commands in marked skid shells,
would lose normal access to existing authentication, configuration, trust,
history and memories. those files remain intact in the existing homes.

source audit, 2026-09-25:

- pre-restoration skid used existing account homes (`927c554` architecture;
  dev-server `c633db5` host config). the fresh homes are a separation decision.
- dev-server `1296309`, `lib/provider-homes.sh`, provisions instructions and
  settings; `assets/skid-provider/shell-init` redirects manual commands too.
- skid's hook admission binds the exact pane tty and foreground process
  lifetime. the native claude helper checks the observed pid and optional
  session id. home separation is not their runtime identity mechanism.
- changing configured paths alone is unsafe: `lib/skid-provider.sh` replaces
  each codex `hooks.json` with the skid file. pointed at an existing home,
  that would discard other hooks. the claude plugin is loaded explicitly.

accepted correction: preserve existing provider homes for skid too. its
`3bd0ae4` removes the unused codex hook writer/template entirely; no merger is
needed. claude retains its explicitly loaded plugin, guarded at provider exec
before reading input/config. personal claude keeps its native unset home
variable. the owner reports engineering and disposable probes passing.
dev-server must adopt this source handoff, remove account provisioning and
codex hook replacement, and qualify actual hook/runtime coexistence.

resolved when the shared-home launch/integration contract preserves ordinary
commands and provider state, and concurrent products are qualified at home
and a shared project. actual provider behavior remains `NOT_RUN`; neither
fresh homes nor source checks alone establish hook isolation.
