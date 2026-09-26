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
repo's `3bd0ae4` handoff and root deployment assignment now require this;
dev-server implementation is pending. preserve normal shell startup and
account selection; do not work around the coupling by replacing user dotfiles.

resolved when ordinary provider maintenance has no skid-specific prerequisite,
and skid's supported login/interactive launch paths still load their own
integration. the dev-server owner owns the implementation and disposable
shell qualification.
