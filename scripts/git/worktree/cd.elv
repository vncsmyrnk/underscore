#!/usr/bin/env elvish

var fzf-flags = '--height=~10'

if (not (has-env UNDERSCORE_IPC_CWD)) {
  echo 'Unset IPC file descriptor.' >&2
  exit 1
}

var wt-path = (
  git worktree list |
    fzf $fzf-flags |
    awk '{ print $1 }'
)

echo $wt-path >$E:UNDERSCORE_IPC_CWD
