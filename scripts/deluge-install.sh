#!/bin/bash

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
