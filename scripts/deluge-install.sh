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
    export DEBIAN_FRONTEND="noninteractive"

    apt-key adv --keyserver keyserver.ubuntu.com --recv-keys C5E6A5ED249AD24C
    sh -c 'echo "deb http://ppa.launchpad.net/deluge-team/stable/ubuntu bionic main" >> /etc/apt/sources.list.d/deluge.list'

    apt-get update

    apt-get install -qq -y deluged
elif [ "$1" = "--v1" ]; then
    apt-get install -qq python-libtorrent python-twisted-core python-openssl python-xdg
    wget https://launchpad.net/ubuntu/+archive/primary/+files/deluge-common_1.3.15-2_all.deb https://launchpad.net/ubuntu/+archive/primary/+files/deluged_1.3.15-2_all.deb
    dpkg -i *.deb
    rm *.deb
else
    echo "ERROR: invalid argument" 1>&2
    exit 2
fi
