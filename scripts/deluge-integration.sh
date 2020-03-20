#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: deluge-install.sh (--v1|--v2)" 1>&2
    exit 1
fi

set -e

if [ "$1" = "--v2" ]; then
    DCLI_ARGS="-v2"
elif [ "$1" = "--v1" ]; then
    DCLI_ARGS=""
else
    echo "ERROR: invalid argument" 1>&2
    exit 2
fi

## create auth file
mkdir -p $HOME/.config/deluge
echo 'localclient:deluge:10' > $HOME/.config/deluge/auth

## start daemon locally
deluged --do-not-daemonize --loglevel info &
trap "kill $!" EXIT

## default password
export DELUGE_PASSWORD="deluge"

## integration tests
chmod +x bin/delugecli

I=0
while ! ss --no-header --listening --numeric --tcp | grep -qF ':58846'; do
  sleep 1
  let I+=1
  if [ $I -eq 60 ]; then
    echo "Failed to wait for daemon" 1>&2
    exit 1
  fi
done

## run all integration tests
bin/delugecli -host 127.0.0.1 -integration-tests $DCLI_ARGS
