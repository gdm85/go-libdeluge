/*
 * go-libdeluge v0.3.1 - a native deluge RPC client library
 * Copyright (C) 2015~2019 gdm85 - https://github.com/gdm85/go-libdeluge/
This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/
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

const (
	DefaultReadWriteTimeout = time.Second * 30
)

var (
	// ErrAlreadyClosed is returned when connection is already closed.
	ErrAlreadyClosed      = errors.New("connection is already closed")
	ErrInvalidListResult  = errors.New("expected dictionary as list response")
	ErrInvalidReturnValue = errors.New("invalid return value")
)

type DelugeClient interface {
	MethodsList() ([]string, error)
	DaemonVersion() (string, error)
	GetFreeSpace(string) (int64, error)
	AddTorrentMagnet(magnetURI string, options Options) (string, error)
	AddTorrentURL(url string, options Options) (string, error)
	DeleteTorrent(id string) (bool, error)
	TorrentsStatus() (map[string]*TorrentStatus, error)
	MoveStorage(torrentIDs []string, dest string) error
	SetTorrentTracker(id, tracker string) error
	SetTorrentOptions(id string, options Options) error
	SessionState() ([]string, error)
	SetLabel(hash, label string) error
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

// Options used when adding a torrent magnet/URL.
// Valid options are:
//
// * add_paused                   (bool)
// * auto_managed                 (bool)
// * download_location            (string)
// * max_connections              (int)
// * max_download_speed           (int)
// * max_upload_slots             (int)
// * max_upload_speed             (int)
// * move_completed               (bool)
// * move_completed_path          (string)
// * pre_allocate_storage         (bool)
// * prioritize_first_last_pieces (bool)
// * remove_at_ratio              (float32)
// * sequential_download          (bool)
// * shared                       (bool)
// * stop_at_ratio                (bool)
// * stop_ratio                   (float32)
// * super_seeding                (bool)
//
// (from  https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/core/torrent.py#L167-L183)
type Options map[string]interface{}

// Settings defines all settings for a Deluge client connection.
type Settings struct {
	Hostname         string
	Port             uint
	Login            string
	Password         string
	Logger           *log.Logger
	ReadWriteTimeout time.Duration // Timeout for read/write operations on the TCP stream.
	// V2Daemon enables the new v1 protocol for v2 daemons.
	V2Daemon              bool
	DebugSaveInteractions bool
}

// Client is a Deluge RPC client.
type Client struct {
	settings      Settings
	safeConn      SafeConn
	serial        int64
	classID       int64
	DebugIncoming [][]byte
}

type SafeConn struct {
	conn             *tls.Conn
	readWriteTimeout time.Duration
}

var _ DelugeClient = &Client{}
var _ NativeDelugeClient = &Client{}

type rpcResponseTypeID int

type TorrentStatus struct {
	NumSeeds            int64
	Ratio               float32
	Progress            float32 // max is 100
	DistributedCopies   float32
	TotalDone           int64
	SeedingTime         int64
	ETA                 float32 // most times an integer
	IsFinished          bool
	NumPieces           int64
	TrackerHost         string
	PieceLength         int64
	ActiveTime          int64
	IsSeed              bool
	NumPeers            int64
	NextAnnounce        int64
	Name                string
	Label               string
	State               string
	TotalSeeds          int64
	TotalPeers          int64
	DownloadPayloadRate int64
	UploadPayloadRate   int64
	TrackerStatus       string
	TotalSize           int64
	DownloadLocation    string

	Files          []File
	Peers          []Peer
	FilePriorities []int64
	FileProgress   []float32
}

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

	// rencode -> zlib -> openssl -> TCP
	var b bytes.Buffer
	z := zlib.NewWriter(&b)
	e := rencode.NewEncoder(z)

	// payload is wrapped twice in a list because there is support for multiple RPC calls
	// although not used currently
	payload := rencode.NewList(rencode.NewList(c.serial, methodName, args, kwargs))

	err := e.Encode(payload)
	if err != nil {
		return nil, err
	}

	// flush zlib-compressed buffer
	err = z.Close()
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Println("flushed zlib buffer")
	}

	// write to connection without closing it
	if c.settings.V2Daemon {
		// on v2+ send the header
		var header [5]byte
		header[0] = PROTOCOL_VERSION
		binary.BigEndian.PutUint32(header[1:], uint32(b.Len()))
		_, err = c.safeConn.Write(header[:])
		if err != nil {
			return nil, err
		}
		if c.settings.Logger != nil {
			c.settings.Logger.Printf("written V2 header %X to RPC connection", header[:])
		}
	}
	n, err := c.safeConn.Write(b.Bytes())
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Printf("written %d bytes to RPC connection", n)
	}

	// setup a reader: TCP -> openssl -> zlib -> (header in V2) rencode -> {objects}
	var src io.Reader
	if !c.settings.V2Daemon {
		src = &c.safeConn
	} else {
		// on v2+ first identify the header, then use the compressed body (more inefficient)
		// a zlib header could be automatically detected but it's pointless since we use a flag to identify V2 daemons
		// (remote endpoint does not version handshakes)
		var header [5]byte
		_, err = c.safeConn.Read(header[:])
		if err != nil {
			return nil, err
		}

		if header[0] != PROTOCOL_VERSION {
			return nil, fmt.Errorf("found protocol version %d but expected %d", header[0], PROTOCOL_VERSION)
		}

		l := binary.BigEndian.Uint32(header[1:])
		b.Reset()
		_, err = io.CopyN(&b, &c.safeConn, int64(l))
		if err != nil {
			return nil, err
		}

		src = &b
	}

	zr, err := zlib.NewReader(src)
	if err != nil {
		return nil, err
	}

	var d *rencode.Decoder
	if c.settings.DebugSaveInteractions {
		var inB bytes.Buffer
		_, err = io.Copy(&inB, zr)
		if err != nil {
			return nil, err
		}
		inBytes := inB.Bytes()
		d = rencode.NewDecoder(bytes.NewReader(inBytes))
		c.DebugIncoming = append(c.DebugIncoming, inBytes)
	} else {
		d = rencode.NewDecoder(zr)
	}

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

	err = enableKeepAlive(rawConn, 10*time.Second, 2, 5*time.Second)
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
	resp, err := c.rpc("daemon.get_method_list", rencode.List{}, rencode.Dictionary{})
	if err != nil {
		return []string{}, err
	}
	if resp.IsError() {
		return []string{}, resp.RPCError
	}

	var methodsList rencode.List
	err = resp.returnValue.Scan(&methodsList)
	if err != nil {
		return []string{}, err
	}
	result := make([]string, methodsList.Length())
	for i, m := range methodsList.Values() {
		result[i] = string(m.([]byte))
	}

	return result, nil
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

func mapToRencodeDictionary(m map[string]interface{}) rencode.Dictionary {
	var dict rencode.Dictionary
	if m != nil {
		for k, v := range m {
			dict.Add(k, v)
		}
	}

	return dict
}

func sliceToRencodeList(s []string) rencode.List {
	var list rencode.List
	for _, v := range s {
		list.Add(v)
	}

	return list
}
