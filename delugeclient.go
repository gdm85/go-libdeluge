// go-libdeluge v0.4.1 - a native deluge RPC client library
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
	DefaultReadWriteTimeout = time.Second * 30
)

var (
	// ErrAlreadyClosed is returned when connection is already closed.
	ErrAlreadyClosed             = errors.New("connection is already closed")
	ErrInvalidDictionaryResponse = errors.New("expected dictionary as list response")
	ErrInvalidReturnValue        = errors.New("invalid return value")
	ErrUnsupportedV1             = errors.New("method not supported by deluge daemon v1")
)

type DelugeClient interface {
	MethodsList() ([]string, error)
	DaemonVersion() (string, error)
	GetFreeSpace(string) (int64, error)
	AddTorrentMagnet(magnetURI string, options *Options) (string, error)
	AddTorrentURL(url string, options *Options) (string, error)
	RemoveTorrents(ids []string, rmFiles bool) ([]TorrentError, error)
	RemoveTorrent(id string, rmFiles bool) (bool, error)
	PauseTorrents(ids []string) error
	PauseTorrent(id string) error
	ResumeTorrents(ids []string) error
	ResumeTorrent(id string) error
	TorrentsStatus(state TorrentState, ids []string) (map[string]*TorrentStatus, error)
	TorrentStatus(id string) (*TorrentStatus, error)
	MoveStorage(torrentIDs []string, dest string) error
	SetTorrentTracker(id, tracker string) error
	SetTorrentOptions(id string, options *Options) error
	SessionState() ([]string, error)
	KnownAccounts() ([]Account, error)
	CreateAccount(account Account) (bool, error)
	UpdateAccount(account Account) (bool, error)
	RemoveAccount(username string) (bool, error)
	GetAvailablePlugins() ([]string, error)
	GetEnabledPlugins() ([]string, error)
	GetAvailablePluginsLookup() (map[string]struct{}, error)
	GetEnabledPluginsLookup() (map[string]struct{}, error)
}

type NativeDelugeClient interface {
	Close() error
	Connect() error
}

type SerialMismatchError struct {
	ExpectedID int64
	ReceivedID int64
}

func (e SerialMismatchError) Error() string {
	return fmt.Sprintf("request/response serial id mismatch: got %d but %d expected", e.ReceivedID, e.ExpectedID)
}

// Settings defines all settings for a Deluge client connection.
type Settings struct {
	Hostname         string
	Port             uint
	Login            string
	Password         string
	Logger           *log.Logger
	ReadWriteTimeout time.Duration // Timeout for read/write operations on the TCP stream.
	// V2Daemon enables the new v1 protocol for v2 daemons.
	V2Daemon bool
	// DebugSaveInteractions is used populate the DebugIncoming slice on the client with
	// byte buffers containing the raw bytes as received from the Deluge server.
	DebugSaveInteractions bool
}

// Client is a Deluge RPC client.
type Client struct {
	settings      Settings
	safeConn      SafeConn
	serial        int64
	classID       int64
	DebugIncoming []*bytes.Buffer
}

type SafeConn struct {
	conn             *tls.Conn
	readWriteTimeout time.Duration
}

var _ DelugeClient = &Client{}
var _ NativeDelugeClient = &Client{}

type rpcResponseTypeID int

type File struct {
	Index  int64
	Size   int64
	Offset int64
	Path   string
}

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
	rpcResponse rpcResponseTypeID = 1
	rpcError    rpcResponseTypeID = 2
	rpcEvent    rpcResponseTypeID = 3
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
	messageType rpcResponseTypeID
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

func (sc *SafeConn) Read(p []byte) (n int, err error) {
	// set deadline
	err = sc.conn.SetReadDeadline(time.Now().Add(sc.readWriteTimeout))
	if err != nil {
		return 0, err
	}

	return sc.conn.Read(p)
}

func (sc *SafeConn) Write(p []byte) (n int, err error) {
	// set deadline
	err = sc.conn.SetWriteDeadline(time.Now().Add(sc.readWriteTimeout))
	if err != nil {
		return 0, err
	}

	return sc.conn.Write(p)
}

// protocol version used with Deluge v2+
const PROTOCOL_VERSION = 1

func (c *Client) rpc(methodName string, args rencode.List, kwargs rencode.Dictionary) (*DelugeResponse, error) {
	if c.safeConn.conn == nil {
		return nil, ErrAlreadyClosed
	}
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
	if c.settings.V2Daemon {
		// on v2+ send the header
		var header [5]byte
		header[0] = PROTOCOL_VERSION
		binary.BigEndian.PutUint32(header[1:], uint32(l))
		_, err = c.safeConn.Write(header[:])
		if err != nil {
			return nil, err
		}
		if c.settings.Logger != nil {
			c.settings.Logger.Printf("V2 request header: %X", header[:])
		}
	}
	n, err := io.Copy(&c.safeConn, &reqBytes)
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
	var src io.Reader = &c.safeConn

	// when debugging copy the source bytes as they are received
	if c.settings.DebugSaveInteractions {
		var copyOfResponseBytes bytes.Buffer
		src = io.TeeReader(src, &copyOfResponseBytes)

		c.DebugIncoming = append(c.DebugIncoming, &copyOfResponseBytes)
	}

	if c.settings.V2Daemon {
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

		if header[0] != PROTOCOL_VERSION {
			return nil, fmt.Errorf("found protocol version %d but expected %d", header[0], PROTOCOL_VERSION)
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

	// buffered input here is necessary to avoid the rencode decoder racing past the ZLib decoder
	var deflatedBuf bytes.Buffer
	_, err = io.Copy(&deflatedBuf, zr)
	if err != nil {
		return nil, err
	}
	d := rencode.NewDecoder(&deflatedBuf)

	resp, err := c.handleRpcResponse(d, c.serial)
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Printf("RPC(%s) = %s\n", methodName, resp.String())
	}
	return resp, nil
}

func (c *Client) handleRpcResponse(d *rencode.Decoder, expectedSerial int64) (*DelugeResponse, error) {
	var respList rencode.List
	err := d.Scan(&respList)
	if err != nil {
		return nil, err
	}

	var resp DelugeResponse
	var mt int64
	err = respList.Scan(&mt, &resp.requestID)
	if err != nil {
		return nil, err
	}
	if resp.requestID != expectedSerial {
		return nil, SerialMismatchError{expectedSerial, resp.requestID}
	}
	resp.messageType = rpcResponseTypeID(mt)

	// shift first two elements which have already been read
	respList = rencode.NewList(respList.Values()[2:]...)

	switch resp.messageType {
	case rpcResponse:
		resp.returnValue = respList
	case rpcError:
		if c.settings.V2Daemon {
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
	case rpcEvent:
		return nil, errors.New("event support not available")
	default:
		return nil, errors.New("unknown message type")
	}

	return &resp, nil
}

// New returns a Deluge client.
func New(s Settings) *Client {
	if s.ReadWriteTimeout == time.Duration(0) {
		s.ReadWriteTimeout = DefaultReadWriteTimeout
	}
	return &Client{
		settings: s,
	}
}

// Close closes the connection of a Deluge client.
func (c *Client) Close() error {
	if c.safeConn.conn == nil {
		return ErrAlreadyClosed
	}
	err := c.safeConn.conn.Close()
	c.safeConn.conn = nil
	return err
}

// Connect performs connection to a Deluge daemon second previously specified settings.
func (c *Client) Connect() error {
	dialer := new(net.Dialer)
	rawConn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", c.settings.Hostname, c.settings.Port))
	if err != nil {
		return err
	}

	c.safeConn.conn = tls.Client(rawConn, &tls.Config{
		ServerName:         c.settings.Hostname,
		InsecureSkipVerify: true, // x509: cannot verify signature: algorithm unimplemented
	})
	c.safeConn.readWriteTimeout = c.settings.ReadWriteTimeout

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("connected to %s:%d\n", c.settings.Hostname, c.settings.Port)
	}

	var kwargs rencode.Dictionary

	// in v2+ the client version must be specified
	if c.settings.V2Daemon {
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
	err = resp.returnValue.Scan(&c.classID)
	if err != nil {
		return err
	}

	if c.settings.Logger != nil {
		c.settings.Logger.Println("login successful as user", c.settings.Login)
	}

	return nil
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
