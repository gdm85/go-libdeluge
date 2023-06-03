#!/bin/bash
## go-libdeluge v0.5.6 - a native deluge RPC client library
## Copyright (C) 2015~2023 gdm85 - https://github.com/gdm85/go-libdeluge/
## This program is free software; you can redistribute it and/or
## modify it under the terms of the GNU General Public License
## as published by the Free Software Foundation; either version 2
## of the License, or (at your option) any later version.
## This program is distributed in the hope that it will be useful,
## but WITHOUT ANY WARRANTY; without even the implied warranty of
## MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
## GNU General Public License for more details.
## You should have received a copy of the GNU General Public License
## along with this program; if not, write to the Free Software
## Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

if [ $# -ne 1 ]; then
    echo "Usage: deluge-install.sh (--v1|--v2)" 1>&2
    exit 1
fi

set -e

if [ "$1" = "--v2" ]; then
    BIN=inttest2
elif [ "$1" = "--v1" ]; then
    BIN=inttest1
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
chmod +x bin/$BIN
bin/$BIN -test.v
