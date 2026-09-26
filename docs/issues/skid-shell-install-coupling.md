# skid shell setup can block shared provider maintenance

problem: dev-server `1296309` adds `ai_install_skid_shell_init` unconditionally
to shared `ai_install`, even when original skid is not installed. it edits
`.bashrc`, the effective bash login file, and an existing `.zlogin`.

impact: a symlinked startup file now returns an action/error from the shared
provider installation solely because of skid support. ordinary maintenance
acquires a prerequisite belonging to an optional product. the runtime guard
does keep skid's command functions inactive in ordinary and herdr shells;
this is installation coupling, not another global provider reroute.

evidence: `lib/ai-tools.sh`, `ai_install` calls the new function and propagates
its failure; `ai_install_skid_shell_source` returns 2 for a symlinked target.
these paths were introduced after baseline `8498933`. source inspection only;
no user startup files were read or modified during this audit.

accepted correction: skid owns its required shell integration. the original
repo's `0bb7e2a` handoff and root deployment assignment require this.
dev-server's working tree removes the shared `ai_install` call, and its owner
reports disposable before/after, symlink, mode, idempotence and late-override
checks passing. final reviewed deployment source remains pending.

bounded trade-off: fully managed `.zshrc` retains an optional source guard,
checking the skid marker and absence of herdr context before any skid file
access. otherwise ordinary whole-file updates erase the integration. this
static dependency neither reroutes ordinary commands nor adds a shared
provider-installation prerequisite. preserve normal startup and account
selection; do not replace user dotfiles to resolve skid-specific requirements.

resolved when ordinary provider maintenance has no skid-specific prerequisite,
and skid's supported login/interactive launch paths still load their own
integration. the dev-server owner owns the implementation and disposable
shell qualification.
