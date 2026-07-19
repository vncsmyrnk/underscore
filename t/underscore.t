#!/usr/bin/env elvish

# vim: set ft=elvish:

use path
use str
use runtime
use github.com/tesujimath/elvish-tap/tap

var root = (path:abs (path:dir (src)[name])'/..')

fn setup-fixture {
  var tmp = (path:temp-dir)
  mkdir -p $tmp'/bin' $tmp'/share/underscore/scripts'
  cp $root'/underscore.elv' $tmp'/bin/underscore.elv'
  chmod 755 $tmp'/bin/underscore.elv'
  put $tmp
}

fn test-executes-matching-script {
  var tmp = (setup-fixture)

  mkdir -p $tmp'/share/underscore/scripts/some/command'
  echo '#!'$runtime:elvish-path >$tmp'/share/underscore/scripts/some/command/here.elv'
  echo 'echo $args[0] >'$tmp'/forwarded.txt' >>$tmp'/share/underscore/scripts/some/command/here.elv'
  chmod 755 $tmp'/share/underscore/scripts/some/command/here.elv'

  var run = ?(elvish $tmp'/bin/underscore.elv' some command here foo bar)
  var forwarded-args = (cat $tmp'/forwarded.txt')

  rm -rf $tmp

  tap:assert-expected $forwarded-args 'foo bar'
  tap:assert-expected $run $ok
}

fn test-no-script-found {
  var tmp = (setup-fixture)
  var run = ?(elvish $tmp'/bin/underscore.elv' not-found >/dev/null 2>$tmp'/stderr.txt')
  var run-stderr = (cat $tmp'/stderr.txt')

  rm -rf $tmp

  tap:assert-expected $run[reason][exit-status] 1
  tap:assert-expected $run-stderr 'No script found.'
}

fn run-tests {
  tap:run [
    [&d='executes matching nested script and forwards remaining args' &f={ test-executes-matching-script }]
    [&d='returns exit code 1 and reports missing script' &f={ test-no-script-found }]
  ]
}

run-tests
