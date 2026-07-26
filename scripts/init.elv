if (and (> (count $args) 0) (eq $args[0] "--help")) {
  echo "Usage: underscore init <SHELL>"
  exit 0
}

if (== (count $args) 0) {
  fail "a shell is required."
}

use path

var share-path = (get-env UNDERSCORE_DATA_DIR)

var shell = $args[0]
if (eq $shell "zsh") {
  exec cat $share-path$path:separator'entrypoints'$path:separator'zsh'
}

fail "invalid shell."
