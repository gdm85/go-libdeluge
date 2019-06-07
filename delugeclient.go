/*
 * go-libdeluge v0.2.0 - a native deluge RPC client library
 * Copyright (C) 2015~2017 gdm85 - https://github.com/gdm85/go-libdeluge/
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
	AddTorrentFile(fileName, fileContentBase64 string, options Options) (string, error)
	AddTorrentURL(url string, options Options) (string, error)
	DeleteTorrent(id string) (bool, error)
	TorrentsStatus() (map[string]*TorrentStatus, error)
	MoveStorage(torrentIDs []string, dest string) error
	SessionState() ([]string, error)
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

// Options used when adding a torrent magnet/URL
type Options map[string]interface{}

// Settings defines all settings for a Deluge client connection.
type Settings struct {
	Hostname              string
	Port                  uint
	Login                 string
	Password              string
	Logger                *log.Logger
	ReadWriteTimeout      time.Duration // Timeout for read/write operations on the TCP stream.
	DebugSaveInteractions bool
}

// Client is a Deluge RPC client.
type Client struct {
	settings      Settings
	conn          *tls.Conn
	serial        int64
	classID       int64
	DebugIncoming [][]byte
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
	State               string
	TotalSeeds          int64
	TotalPeers          int64
	DownloadPayloadRate int64
	UploadPayloadRate   int64
	TrackerStatus       string
	TotalSize           int64

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
	return fmt.Sprintf("RPC error %s('%s')\n%s", e.ExceptionType, e.ExceptionMessage, e.TraceBack)
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

func (c *Client) resetTimeout() error {
	// set timeout
	return c.conn.SetDeadline(time.Now().Add(c.settings.ReadWriteTimeout))
}

func (c *Client) rpc(methodName string, args rencode.List, kwargs rencode.Dictionary) (*DelugeResponse, error) {
	if c.conn == nil {
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
	var n int
	n, err = c.conn.Write(b.Bytes())
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		//		c.settings.Logger.Println(hex.Dump(b.Bytes()))
		c.settings.Logger.Printf("written %d bytes to RPC connection", n)
	}

	err = c.resetTimeout()
	if err != nil {
		return nil, err
	}

	// setup a reader: TCP -> openssl -> zlib -> rencode -> {objects}
	zr, err := zlib.NewReader(c.conn)
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

	resp, err := handleRpcResponse(d, c.serial)
	if err != nil {
		return nil, err
	}
	if c.settings.Logger != nil {
		c.settings.Logger.Printf("RPC(%s) = %s\n", methodName, resp.String())
	}
	return resp, nil
}

func handleRpcResponse(d *rencode.Decoder, expectedSerial int64) (*DelugeResponse, error) {
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

	// shift first two elements
	respList = rencode.NewList(respList.Values()[2:]...)

	switch resp.messageType {
	case rpcResponse:
		resp.returnValue = respList
	case rpcError:
		var errList rencode.List
		err = respList.Scan(&errList)
		if err != nil {
			return nil, err
		}
		err = errList.Scan(&resp.ExceptionType, &resp.ExceptionMessage, &resp.TraceBack)
		if err != nil {
			return nil, err
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
	if c.conn == nil {
		return ErrAlreadyClosed
	}
	err := c.conn.Close()
	c.conn = nil
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

	c.conn = tls.Client(rawConn, &tls.Config{
		ServerName:         c.settings.Hostname,
		InsecureSkipVerify: true, // x509: cannot verify signature: algorithm unimplemented
	})

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("connected to %s:%d\n", c.settings.Hostname, c.settings.Port)
	}

	// perform login
	resp, err := c.rpc("daemon.login", rencode.NewList(c.settings.Login, c.settings.Password), rencode.Dictionary{})
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

// GetFreeSpace returns the available free space; path is optional.
func (c *Client) GetFreeSpace(path string) (int64, error) {
	var args rencode.List
	args.Add(path)

	resp, err := c.rpc("core.get_free_space", args, rencode.Dictionary{})
	if err != nil {
		return 0, err
	}
	if resp.IsError() {
		return 0, resp.RPCError
	}

	var freeSpace int64
	err = resp.returnValue.Scan(&freeSpace)
	if err != nil {
		return 0, err
	}

	return freeSpace, nil
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

func (c *Client) AddTorrentFile(fileName, fileContentBase64 string, options Options) (string, error) {
	var args rencode.List
	args.Add(fileName, fileContentBase64, mapToRencodeDictionary(options))

	resp, err := c.rpc("core.add_torrent_file", args, rencode.Dictionary{})
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", resp.RPCError
	}

	// returned hash may be nil if torrent was already added
	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return "", ErrInvalidReturnValue
	}
	torrentHash := vals[0]
	//TODO: is this nil comparison valid?
	if torrentHash == nil {
		return "", nil
	}
	return string(torrentHash.([]uint8)), nil
}

// AddTorrentMagnet adds a torrent via magnet URI and returns the torrent hash.
func (c *Client) AddTorrentMagnet(magnetURI string, options Options) (string, error) {
	var args rencode.List
	args.Add(magnetURI, mapToRencodeDictionary(options))

	resp, err := c.rpc("core.add_torrent_magnet", args, rencode.Dictionary{})
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", resp.RPCError
	}

	// returned hash may be nil if torrent was already added
	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return "", ErrInvalidReturnValue
	}
	torrentHash := vals[0]
	//TODO: is this nil comparison valid?
	if torrentHash == nil {
		return "", nil
	}
	return string(torrentHash.([]uint8)), nil
}

// AddTorrentURL adds a torrent via a URL and returns the torrent hash.
func (c *Client) AddTorrentURL(url string, options Options) (string, error) {
	var args rencode.List
	args.Add(url, mapToRencodeDictionary(options))

	resp, err := c.rpc("core.add_torrent_url", args, rencode.Dictionary{})
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", resp.RPCError
	}

	// returned hash may be nil if torrent was already added
	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return "", ErrInvalidReturnValue
	}
	torrentHash := vals[0]
	//TODO: is this nil comparison valid?
	if torrentHash == nil {
		return "", nil
	}
	return string(torrentHash.([]uint8)), nil
}

var STATUS_KEYS = rencode.NewList(
	"state",
	"download_location",
	"tracker_host",
	"tracker_status",
	"next_announce",
	"name",
	"total_size",
	"progress",
	"num_seeds",
	"total_seeds",
	"num_peers",
	"total_peers",
	"eta",
	"download_payload_rate",
	"upload_payload_rate",
	"ratio",
	"distributed_copies",
	"num_pieces",
	"piece_length",
	"total_done",
	"files",
	"file_priorities",
	"file_progress",
	"peers",
	"is_seed",
	"is_finished",
	"active_time",
	"seeding_time")

func (c *Client) TorrentsStatus() (map[string]*TorrentStatus, error) {
	var args rencode.List
	args.Add(rencode.Dictionary{})
	args.Add(STATUS_KEYS)

	resp, err := c.rpc("core.get_torrents_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	return decodeTorrentsStatusResponse(resp)
}

func decodeTorrentsStatusResponse(resp *DelugeResponse) (map[string]*TorrentStatus, error) {
	values := resp.returnValue.Values()
	if len(values) != 1 {
		return nil, ErrInvalidReturnValue
	}
	rd, ok := values[0].(rencode.Dictionary)
	if !ok {
		return nil, ErrInvalidListResult
	}

	d, err := rd.Zip()
	if err != nil {
		return nil, err
	}

	result := map[string]*TorrentStatus{}
	for k, rv := range d {
		v, ok := rv.(rencode.Dictionary)
		if !ok {
			return nil, ErrInvalidListResult
		}

		var ts TorrentStatus
		err = v.ToStruct(&ts)
		if err != nil {
			return nil, err
		}
		result[k] = &ts
	}

	return result, nil
}

func (c *Client) DeleteTorrent(id string) (bool, error) {
	var args rencode.List
	args.Add(id, true)

	// perform login
	resp, err := c.rpc("core.remove_torrent", args, rencode.Dictionary{})
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, resp.RPCError
	}

	// returned hash may be nil if torrent was already added
	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return false, ErrInvalidReturnValue
	}
	success := vals[0]

	return success.(bool), nil
}

func (c *Client) MoveStorage(torrentIDs []string, dest string) error {
	var args rencode.List
	args.Add(sliceToRencodeList(torrentIDs), dest)

	resp, err := c.rpc("core.move_storage", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return err
}

func (c *Client) SessionState() ([]string, error) {
	resp, err := c.rpc("core.get_session_state", rencode.List{}, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	var idList rencode.List
	err = resp.returnValue.Scan(&idList)
	if err != nil {
		return []string{}, err
	}
	result := make([]string, idList.Length())
	for i, m := range idList.Values() {
		result[i] = string(m.([]byte))
	}

	return result, nil
}
