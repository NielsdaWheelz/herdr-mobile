# original skid's existing-account behavior awaits live qualification

problem: original skid and dev-server `ae70f2b` now agree on existing accounts
and scoped hooks. actual provider/history continuity and concurrent runtime
behavior remain unqualified; installer fixtures cannot establish them.

impact of the withdrawn design: skid launches, including manually typed
provider commands, would have hidden existing auth/config/history behind
fresh homes. the corrected design preserves access to those existing accounts.

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
dev-server `ae70f2b` adopts this handoff and removes account provisioning and
codex hook replacement. root confirmed six deployment templates match app
source and shared `ai-tools.sh` matches its pre-separation baseline. the owner
reports 21 account sentinels preserved through the real provider callback,
five command routes and startup/hook-boundary probes passing. actual
authenticated hook/runtime coexistence still needs qualification.

resolved when the shared-home launch/integration contract preserves ordinary
commands and provider state, and concurrent products are qualified at home
and a shared project. actual provider behavior remains `NOT_RUN`; neither
fresh homes nor source checks alone establish hook isolation.
