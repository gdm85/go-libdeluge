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

package delugeclient

import (
	"github.com/gdm85/go-rencode"
)

type LabelPlugin struct {
	*Client
}

// SetTorrentLabel adds or replaces the label for the specified torrent.
func (p LabelPlugin) SetTorrentLabel(hash, label string) error {
	var args rencode.List
	args.Add(hash, label)

	resp, err := p.rpc("label.set_torrent", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return nil
}

// GetTorrentLabel returns the label of the specified torrent.
func (c *Client) GetTorrentLabel(hash string) (string, error) {
	var args rencode.List
	args.Add(hash)
	args.Add(rencode.NewList("label"))

	resp, err := c.rpc("core.get_torrent_status", args, rencode.Dictionary{})
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", resp.RPCError
	}

	rd, err := decodeListWithOneDictionary(resp)
	if err != nil {
		return "", err
	}

	var s struct {
		Label string
	}
	err = rd.ToStruct(&s)
	if err != nil {
		return "", err
	}

	return s.Label, nil
}

// GetTorrentsLabels filters torrents by state and/or IDs and returns their label.
func (c *Client) GetTorrentsLabels(state TorrentState, ids []string) (map[string]string, error) {
	var args rencode.List
	var filterDict rencode.Dictionary
	if len(ids) != 0 {
		filterDict.Add("id", sliceToRencodeList(ids))
	}
	if state != StateUnspecified {
		filterDict.Add("state", string(state))
	}
	args.Add(filterDict)
	args.Add(rencode.NewList("label"))

	resp, err := c.rpc("core.get_torrents_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	rd, err := decodeListWithOneDictionary(resp)
	if err != nil {
		return nil, err
	}

	d, err := rd.Zip()
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for k, rv := range d {
		v, ok := rv.(rencode.Dictionary)
		if !ok {
			return nil, ErrInvalidDictionaryResponse
		}

		var s struct {
			Label string
		}
		err = v.ToStruct(&s)
		if err != nil {
			return nil, err
		}
		result[k] = s.Label
	}

	return result, nil
}
