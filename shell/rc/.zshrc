# vim: set ft=zsh:

\. "$PWD/shell/entrypoint/zsh"

autoload -Uz compinit
compinit

\. "$PWD/completions/zsh"
