#!/usr/bin/env elvish

# Restrict path

use str
use path

var scripts-path = (path:dir (path:abs (src)[name]))'/../share/underscore/scripts'

var i = 1
var script-path = $scripts-path
for arg $args {
  set script-path = $script-path'/'$arg
  if ?(test -x $script-path'.elv') {
    exec $script-path'.elv' (str:join " " $args[$i..])
  }
  set i = (+ $i 1)
}

echo "No script found." >&2
exit 1
