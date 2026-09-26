# original skid's private homes unnecessarily fork provider state

problem: the separation spec still gives original skid five fresh provider
homes. the correction for ordinary/herdr commands leaves this scoped version
of the same disruption intact. separate homes have not been established as a
requirement for independent apps.

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

recommendation: preserve existing provider homes for skid too; isolate owned
runtime state and integrations. the original-app and dev-server owners must
first establish hook coexistence without replacing user/herdr settings.
this finding does not silently change their current implementation contract.

resolved when the shared-home launch/integration contract preserves ordinary
commands and provider state, and concurrent products are qualified at home
and a shared project. actual provider behavior remains `NOT_RUN`. alternatively,
document a demonstrated provider limitation and an explicitly accepted bounded
trade-off. fresh homes alone do not prove hook isolation.
