// go-libdeluge v0.5.4 - a native deluge RPC client library
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

// Package delugeclient allows calling native RPC methods on a remote
// deluge server.
package delugeclient

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/gdm85/go-rencode"
)

// AuthLevel is an Auth Level string understood by Deluge
type AuthLevel string

// The auth level names, as defined in
// https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/core/authmanager.py#L33-L37
const (
	AuthLevelNone     AuthLevel = "NONE"
	AuthLevelReadonly AuthLevel = "READONLY"
	AuthLevelNormal   AuthLevel = "NORMAL"
	AuthLevelAdmin    AuthLevel = "ADMIN"
	AuthLevelDefault  AuthLevel = AuthLevelNormal
)

const (
	// DefaultReadWriteTimeout is the default timeout for I/O operations with the Deluge server.
	DefaultReadWriteTimeout = time.Second * 30
)

var (
	// ErrAlreadyClosed is returned when connection is already closed.
	ErrAlreadyClosed = errors.New("connection is already closed")
	// ErrInvalidDictionaryResponse is returned when the expected dictionary as list is not received.
	ErrInvalidDictionaryResponse = errors.New("expected dictionary as list response")
	// ErrInvalidReturnValue is returned when the returned value received from server is invalid.
	ErrInvalidReturnValue = errors.New("invalid return value")
)

// DelugeClient is an interface for v1.3 and v2 Deluge servers.
type DelugeClient interface {
	Connect() error
	Close() error

	DaemonLogin() error
	MethodsList() ([]string, error)
	DaemonVersion() (string, error)
	GetFreeSpace(string) (int64, error)
	GetLibtorrentVersion() (string, error)
	AddTorrentMagnet(magnetURI string, options *Options) (string, error)
	AddTorrentURL(url string, options *Options) (string, error)
	AddTorrentFile(fileName, fileContentBase64 string, options *Options) (string, error)
	RemoveTorrents(ids []string, rmFiles bool) ([]TorrentError, error)
	RemoveTorrent(id string, rmFiles bool) (bool, error)
	PauseTorrents(ids ...string) error
	ResumeTorrents(ids ...string) error
	TorrentsStatus(state TorrentState, ids []string) (map[string]*TorrentStatus, error)
	TorrentStatus(id string) (*TorrentStatus, error)
	MoveStorage(torrentIDs []string, dest string) error
	SetTorrentTracker(id, tracker string) error
	SetTorrentOptions(id string, options *Options) error
	SessionState() ([]string, error)
	ForceReannounce(ids []string) error
	GetAvailablePlugins() ([]string, error)
	GetEnabledPlugins() ([]string, error)
	GetListenPort() (uint16, error)
	TestListenPort() (bool, error)
	SessionStats(keys ...string) (*SessionStats, error)
}

// V2 is an interface for v2 Deluge clients.
type V2 interface {
	DelugeClient

	KnownAccounts() ([]Account, error)
	CreateAccount(account Account) (bool, error)
	RemoveAccount(username string) (bool, error)
	UpdateAccount(account Account) (bool, error)
}

// Client is a Deluge RPC client.
type Client struct {
	settings     Settings
	safeConn     io.ReadWriteCloser
	serial       int64
	classID      int64
	v2daemon     bool
	excludeV2tag string

	DebugServerResponses []*bytes.Buffer
}

type ClientV2 struct {
	Client
}

var _ DelugeClient = &Client{}
var _ DelugeClient = &ClientV2{}
var _ V2 = &ClientV2{}

// SerialMismatchError is the error returned when server replied with an out-of-order response.
type SerialMismatchError struct {
	ExpectedID int64
	ReceivedID int64
}

func (e SerialMismatchError) Error() string {
	return fmt.Sprintf("request/response serial id mismatch: got %d but %d expected", e.ReceivedID, e.ExpectedID)
}

// Settings defines all settings for a Deluge client connection.
type Settings struct {
	Hostname string
	Port     uint
	Login    string
	Password string
	Logger   *log.Logger
	// ReadWriteTimeout is the timeout for read/write operations on the TCP stream.
	ReadWriteTimeout time.Duration
	// DebugServerResponses is used populate the DebugServerResponses slice on the client with
	// byte buffers containing the raw bytes as received from the Deluge server.
	DebugServerResponses bool
}

type safeConn struct {
	conn             *tls.Conn
	readWriteTimeout time.Duration
}

func newSafeConn(rawConn net.Conn, hostname string, readWriteTimeout time.Duration) *safeConn {
	var sc safeConn
	sc.conn = tls.Client(rawConn, &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true, // x509: cannot verify signature: algorithm unimplemented
	})
	sc.readWriteTimeout = readWriteTimeout
	return &sc
}

type rpcMessageType int

// File is a Deluge torrent file.
type File struct {
	Index  int64
	Size   int64
	Offset int64
	Path   string
}

// Peer is a Deluge torrent peer.
type Peer struct {
	Client    string
	IP        string
	Progress  float32
	Seed      int64
	DownSpeed int64
	UpSpeed   int64
	Country   string
}

const (
	rpcResponse rpcMessageType = 1
	rpcError    rpcMessageType = 2
	rpcEvent    rpcMessageType = 3
)

// RPCError is an error returned by RPC calls.
type RPCError struct {
	ExceptionType    string
	ExceptionMessage string
	TraceBack        string
}

func (e RPCError) Error() string {
	return fmt.Sprintf("RPC error %s('%s')\nTraceback: %s", e.ExceptionType, e.ExceptionMessage, e.TraceBack)
}

// DelugeResponse is a response returned from a completed RPC call.
type DelugeResponse struct {
	messageType rpcMessageType
	requestID   int64
	// only for rpcResponse
	returnValue rencode.List
	// only in rpcError
	RPCError
	// only in rpcEvent
	eventName string
	data      rencode.List
}

// IsError returns true when the response is an error.
func (dr *DelugeResponse) IsError() bool {
	return dr.messageType == rpcError
}

func (dr *DelugeResponse) String() string {
	switch dr.messageType {
	case rpcError:
		return dr.RPCError.Error()
	case rpcResponse:
		typeStr := ""
		for _, v := range dr.returnValue.Values() {
			typeStr += fmt.Sprintf("%T, ", v)
		}
		return fmt.Sprintf("%d return values [%s]", dr.returnValue.Length(), typeStr)
	}
	return fmt.Sprintf("invalid message type: %d", dr.messageType)
}

func (sc *safeConn) Read(p []byte) (n int, err error) {
	// set deadline
	err = sc.conn.SetReadDeadline(time.Now().Add(sc.readWriteTimeout))
	if err != nil {
		return 0, err
	}

	return sc.conn.Read(p)
}

func (sc *safeConn) Write(p []byte) (n int, err error) {
	// set deadline
	err = sc.conn.SetWriteDeadline(time.Now().Add(sc.readWriteTimeout))
	if err != nil {
		return 0, err
	}

	return sc.conn.Write(p)
}

func (sc *safeConn) Close() error {
	if sc.conn == nil {
		return ErrAlreadyClosed
	}
	err := sc.conn.Close()
	sc.conn = nil
	return err
}

// NewV1 returns a Deluge client for v1.3 servers.
func NewV1(s Settings) *Client {
	if s.ReadWriteTimeout == time.Duration(0) {
		s.ReadWriteTimeout = DefaultReadWriteTimeout
	}
	return &Client{
		settings:     s,
		excludeV2tag: "v2only",
	}
}

// NewV2 returns a Deluge client for v1.3 servers.
func NewV2(s Settings) *ClientV2 {
	if s.ReadWriteTimeout == time.Duration(0) {
		s.ReadWriteTimeout = DefaultReadWriteTimeout
	}
	return &ClientV2{
		Client: Client{
			v2daemon: true,
			settings: s,
		},
	}
}

// Close closes the connection of a Deluge client.
func (c *Client) Close() error {
	if c.safeConn == nil {
		return nil
	}
	return c.safeConn.Close()
}

// Deluge2ProtocolVersion is the protocol version used with Deluge v2+
const Deluge2ProtocolVersion = 1

func (c *Client) rpc(methodName string, args rencode.List, kwargs rencode.Dictionary) (*DelugeResponse, error) {
	// generate serial
	c.serial++
	if c.serial == math.MaxInt64 {
		c.serial = 1
	}

	// {Python objects} -> rencode -> ZLib -> openSSL -> TCP
	// the rencode and ZLib steps are covered here
	var reqBytes bytes.Buffer
	zReq := zlib.NewWriter(&reqBytes)
	eReq := rencode.NewEncoder(zReq)

	// payload is wrapped twice in a list because there is support for multiple RPC calls
	// (although not currently used)
	payload := rencode.NewList(rencode.NewList(c.serial, methodName, args, kwargs))

	err := eReq.Encode(payload)
	if err != nil {
		return nil, err
	}

	// flush zlib-compressed buffer
	err = zReq.Close()
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Println("flushed zlib buffer")
	}
	l := reqBytes.Len()

	// write to connection without closing it
	if c.v2daemon {
		// on v2+ send the header
		var header [5]byte
		header[0] = Deluge2ProtocolVersion
		binary.BigEndian.PutUint32(header[1:], uint32(l))
		_, err = c.safeConn.Write(header[:])
		if err != nil {
			return nil, err
		}
		if c.settings.Logger != nil {
			c.settings.Logger.Printf("V2 request header: %X", header[:])
		}
	}
	n, err := io.Copy(c.safeConn, &reqBytes)
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Printf("written %d bytes to RPC connection", n)
	}
	if int(n) != l {
		return nil, fmt.Errorf("expected to write %d raw request bytes but written %d bytes instead", l, n)
	}

	// setup a reader pipeline for the response: TCP -> openssl -> ZLib -> (header in V2) rencode -> {Python objects}
	var src io.Reader = c.safeConn

	// when debugging copy the source bytes as they are received
	if c.settings.DebugServerResponses {
		var copyOfResponseBytes bytes.Buffer
		src = io.TeeReader(src, &copyOfResponseBytes)

		c.DebugServerResponses = append(c.DebugServerResponses, &copyOfResponseBytes)
	}

	if c.v2daemon {
		// on v2+ first identify the header, then use the compressed body (more inefficient)
		// a zlib header could be automatically detected but it's pointless since we use a flag to identify V2 daemons
		// (remote endpoint does not version handshakes)
		var header [5]byte
		_, err = c.safeConn.Read(header[:])
		if err != nil {
			return nil, err
		}
		if c.settings.Logger != nil {
			c.settings.Logger.Printf("V2 response header: %X", header[:])
		}

		if header[0] != Deluge2ProtocolVersion {
			return nil, fmt.Errorf("found protocol version %d but expected %d", header[0], Deluge2ProtocolVersion)
		}

		// read all the advertised bytes at once
		l := binary.BigEndian.Uint32(header[1:])
		var respBytes bytes.Buffer

		n, err := io.CopyN(&respBytes, src, int64(l))
		if err != nil {
			return nil, err
		}

		if n != int64(l) {
			return nil, fmt.Errorf("expected %d bytes read but got %d", l, n)
		}

		src = &respBytes
	}

	zr, err := zlib.NewReader(src)
	if err != nil {
		return nil, err
	}

	d := rencode.NewDecoder(zr)

	resp, err := c.handleRPCResponse(d, c.serial)
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Printf("RPC(%s) = %s\n", methodName, resp.String())
	}
	return resp, nil
}

func (c *Client) handleRPCResponse(d *rencode.Decoder, expectedSerial int64) (*DelugeResponse, error) {
	var respList rencode.List
	err := d.Scan(&respList)
	if err != nil {
		return nil, err
	}

	var resp DelugeResponse
	var mt int64

	err = respList.Scan(&mt)
	if err != nil {
		return nil, err
	}
	respList.Shift(1)
	resp.messageType = rpcMessageType(mt)
	if resp.messageType == rpcEvent {
		err = respList.Scan(&resp.eventName, &resp.data)
		if err != nil {
			return nil, err
		}

		return nil, errors.New("event support not available")
	}

	// start reading request ID (for both valid response or error)
	err = respList.Scan(&resp.requestID)
	if err != nil {
		return nil, err
	}
	respList.Shift(1)
	if resp.requestID != expectedSerial {
		return nil, SerialMismatchError{expectedSerial, resp.requestID}
	}

	switch resp.messageType {
	case rpcResponse:
		resp.returnValue = respList
	case rpcError:
		if c.v2daemon {
			var exceptionArgs rencode.List
			var errDict rencode.Dictionary
			err = respList.Scan(&resp.ExceptionType, &exceptionArgs, &errDict, &resp.TraceBack)
			if err != nil {
				return nil, err
			}
			if exceptionArgs.Length() != 0 {
				v := exceptionArgs.Values()[0]
				if v, ok := v.([]byte); ok {
					resp.ExceptionMessage = string(v)
				}
			}
		} else {
			var errList rencode.List
			err = respList.Scan(&errList)
			if err != nil {
				return nil, err
			}
			err = errList.Scan(&resp.ExceptionType, &resp.ExceptionMessage, &resp.TraceBack)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, errors.New("unknown message type")
	}

	return &resp, nil
}

// Connect performs connection to a Deluge daemon and logs in.
func (c *Client) Connect() error {
	dialer := new(net.Dialer)
	rawConn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", c.settings.Hostname, c.settings.Port))
	if err != nil {
		return err
	}

	c.safeConn = newSafeConn(rawConn, c.settings.Hostname, c.settings.ReadWriteTimeout)

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("connected to %s:%d\n", c.settings.Hostname, c.settings.Port)
	}

	err = c.DaemonLogin()
	if err != nil {
		return err
	}

	if c.settings.Logger != nil {
		c.settings.Logger.Println("login successful as user", c.settings.Login)
	}

	return nil
}

// DaemonLogin performs login to the Deluge daemon.
func (c *Client) DaemonLogin() error {
	var kwargs rencode.Dictionary

	// in v2+ the client version must be specified
	if c.v2daemon {
		kwargs.Add("client_version", "2.0.3")
	}

	// perform login
	resp, err := c.rpc("daemon.login", rencode.NewList(c.settings.Login, c.settings.Password), kwargs)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	// get class of logged-in user
	return resp.returnValue.Scan(&c.classID)
}

// MethodsList returns a list of available methods on server.
func (c *Client) MethodsList() ([]string, error) {
	return c.rpcWithStringsResult("daemon.get_method_list")
}

func (c *Client) rpcWithStringsResult(method string) ([]string, error) {
	resp, err := c.rpc(method, rencode.List{}, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	var list rencode.List
	err = resp.returnValue.Scan(&list)
	if err != nil {
		return nil, err
	}
	result := make([]string, list.Length())
	for i, v := range list.Values() {
		result[i] = string(v.([]byte))
	}

	return result, nil
}

func (c *Client) rpcWithDictionaryResult(methodName string, args rencode.List, kwargs rencode.Dictionary) (rencode.Dictionary, error) {
	var (
		rd rencode.Dictionary
		ok bool
	)
	resp, err := c.rpc(methodName, args, kwargs)
	if err != nil {
		return rd, err
	}
	if resp.IsError() {
		return rd, resp.RPCError
	}

	values := resp.returnValue.Values()
	if len(values) != 1 {
		return rd, ErrInvalidReturnValue
	}
	rd, ok = values[0].(rencode.Dictionary)
	if !ok {
		return rd, ErrInvalidDictionaryResponse
	}

	return rd, nil
}

// DaemonVersion returns the running daemon version.
func (c *Client) DaemonVersion() (string, error) {
	resp, err := c.rpc("daemon.info", rencode.List{}, rencode.Dictionary{})
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", resp.RPCError
	}

	var info string
	err = resp.returnValue.Scan(&info)
	if err != nil {
		return "", err
	}

	return info, nil
}
