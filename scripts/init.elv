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
  var zsh-rc = $share-path$path:separator'shell'$path:separator'entrypoint'$path:separator'zsh'
  exec cat $zsh-rc
}

fail "invalid shell."
