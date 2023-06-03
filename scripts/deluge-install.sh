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

    ## NOTE: geoip-database is optional
    apt-get update
    apt-get install -qq -y deluged geoip-database python3-pip

    ## necessary due to https://github.com/ratanakvlun/deluge-ltconfig/issues/23
    cd /usr/lib/python3/dist-packages
    cat<<EOF | patch -p1
diff --git a/deluge/log.py b/deluge/log.py
index 75e8308b5..0f9877fdb 100644
--- a/deluge/log.py
+++ b/deluge/log.py
@@ -86,7 +86,7 @@ class Logging(LoggingLoggerClass):
     def exception(self, msg, *args, **kwargs):
         yield LoggingLoggerClass.exception(self, msg, *args, **kwargs)
 
-    def findCaller(self, stack_info=False):  # NOQA: N802
+    def findCaller(self, *args, **kwargs):  # NOQA: N802
         f = logging.currentframe().f_back
         rv = '(unknown file)', 0, '(unknown function)'
         while hasattr(f, 'f_code'):
EOF
elif [ "$1" = "--v1" ]; then
    ## NOTE: geoip-database is optional
    apt-get install -qq wget geoip-database

    ## based on https://github.com/josh-gaby/deluge1.3.15-Ubuntu-20.04
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/libt/libtorrent-rasterbar/python-libtorrent_1.1.5-1build1_amd64.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/libt/libtorrent-rasterbar/libtorrent-rasterbar9_1.1.5-1build1_amd64.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/main/b/boost1.65.1/libboost-system1.65.1_1.65.1+dfsg-0ubuntu5_amd64.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/b/boost1.65.1/libboost-python1.65.1_1.65.1+dfsg-0ubuntu5_amd64.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/d/deluge/deluge-common_1.3.15-2_all.deb
    wget -q --show-progress --progress=bar:force http://security.ubuntu.com/ubuntu/pool/main/t/twisted/python-twisted-core_17.9.0-2ubuntu0.3_all.deb
    wget -q --show-progress --progress=bar:force http://security.ubuntu.com/ubuntu/pool/main/t/twisted/python-twisted-bin_17.9.0-2ubuntu0.3_amd64.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/main/i/incremental/python-incremental_16.10.1-3_all.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/d/deluge/deluge-common_1.3.15-2_all.deb
    wget -q --show-progress --progress=bar:force http://archive.ubuntu.com/ubuntu/pool/universe/d/deluge/deluged_1.3.15-2_all.deb

    apt install -qq -y ./*.deb
    rm *.deb
else
    echo "ERROR: invalid argument" 1>&2
    exit 2
fi
