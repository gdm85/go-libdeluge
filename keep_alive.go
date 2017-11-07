package delugeclient

import (
	"fmt"
	"net"
	"os"
	"syscall"
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

func setIdle(fd int, secs int) error {
	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs))
}

func setCount(fd int, n int) error {
	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, n))
}

func setInterval(fd int, secs int) error {
	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs))
}

func secs(d time.Duration) int {
	d += (time.Second - time.Nanosecond)
	return int(d.Seconds())
}
