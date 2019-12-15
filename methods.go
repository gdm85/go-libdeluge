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
	"github.com/gdm85/go-rencode"
)

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
	"label",
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

// SetLabel adds or replaces a label for a torrent with given hash
func (c *Client) SetLabel(hash, label string) error {
	var args rencode.List
	args.Add(hash, label)

	resp, err := c.rpc("label.set_torrent", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return nil
}
