// go-libdeluge v0.4.0 - a native deluge RPC client library
// Copyright (C) 2015~2020 gdm85 - https://github.com/gdm85/go-libdeluge/
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

package delugeclient

import (
	"fmt"
	"net"
	"time"
)

// enableKeepAlive enables TCP keepalive for the given conn, which must be a
// *tcp.TCPConn. The returned Conn allows overwriting the default keepalive
// parameters used by the operating system.
// See also: http://felixge.de/2014/08/26/tcp-keepalive-with-golang.html
func enableKeepAlive(conn net.Conn, idleTime time.Duration, count int, interval time.Duration) error {
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("Bad conn type: %T", conn)
	}
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}
	f, err := tcp.File()
	if err != nil {
		return err
	}
	defer f.Close()

	fd := int(f.Fd())
	if err = setIdle(fd, secs(idleTime)); err != nil {
		return err
	}

	if err = setCount(fd, count); err != nil {
		return err
	}

	if err = setInterval(fd, secs(interval)); err != nil {
		return err
	}

	/*	if err = setNonblock(fd); err != nil {
		return err
	}*/

	return nil
}

func secs(d time.Duration) int {
	d += (time.Second - time.Nanosecond)
	return int(d.Seconds())
}
