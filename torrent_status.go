// go-libdeluge v0.5.6 - a native deluge RPC client library
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

// TorrentStatus contains commonly used torrent attributes, as reported
// by the deluge server.
// The full list of potentially available attributes can be found here:
// https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/core/torrent.py#L1033-L1143
// If a new field is added to this struct it should also be added to the statusKeys map.
type TorrentStatus struct {
	ActiveTime          int64
	CompletedTime       int64   `rencode:"v2only"`
	TimeAdded           float32 // most times an integer
       LastSeenComplete    float32
	DistributedCopies   float32
	ETA                 float32 // most times an integer
	Progress            float32 // max is 100
	Ratio               float32
	IsFinished          bool
	IsSeed              bool
	Private             bool
	DownloadLocation    string `rencode:"v2only"`
	DownloadPayloadRate int64
	Name                string
	NextAnnounce        int64
	NumPeers            int64
	NumPieces           int64
	NumSeeds            int64
	PieceLength         int64
	SeedingTime         int64
	State               string
	TotalDone           int64
	TotalPeers          int64
	TotalSeeds          int64
	TotalSize           int64
	TrackerHost         string
	TrackerStatus       string
	UploadPayloadRate   int64

	Files          []File
	Peers          []Peer
	FilePriorities []int64
	FileProgress   []float32
}

type TorrentState string

// See all defined torrent states here: https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/common.py#L70-L78
// Plus the special 'Active' state.
const (
	StateUnspecified TorrentState = ""
	StateActive      TorrentState = "Active"
	StateAllocating  TorrentState = "Allocating"
	StateChecking    TorrentState = "Checking"
	StateDownloading TorrentState = "Downloading"
	StateSeeding     TorrentState = "Seeding"
	StatePaused      TorrentState = "Paused"
	StateError       TorrentState = "Error"
	StateQueued      TorrentState = "Queued"
	StateMoving      TorrentState = "Moving"
)

// each of the available fields in a torrent status
// fields differ from v1/v2
// See current list at https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/core/torrent.py#L1033-L1143
var statusKeys = rencode.NewList(
	"state",
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
	"seeding_time",
	"time_added",
	"completed_time",    // v2-only
	"download_location", // v2-only
       "last_seen_complete",
	"private")

// TorrentStatus returns the status of the torrent with specified hash.
func (c *Client) TorrentStatus(hash string) (*TorrentStatus, error) {
	var args rencode.List
	args.Add(hash)
	args.Add(statusKeys)

	rd, err := c.rpcWithDictionaryResult("core.get_torrent_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}

	var ts TorrentStatus
	err = rd.ToStruct(&ts, c.excludeV2tag)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

// TorrentsStatus returns the status of torrents matching the specified state and list of hashes.
// Both state and list of hashes are optional.
func (c *Client) TorrentsStatus(state TorrentState, hashes []string) (map[string]*TorrentStatus, error) {
	var args rencode.List
	var filterDict rencode.Dictionary
	if len(hashes) != 0 {
		filterDict.Add("id", sliceToRencodeList(hashes))
	}
	if state != StateUnspecified {
		filterDict.Add("state", string(state))
	}
	args.Add(filterDict)
	args.Add(statusKeys)

	rd, err := c.rpcWithDictionaryResult("core.get_torrents_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}

	d, err := rd.Zip()
	if err != nil {
		return nil, err
	}

	result := map[string]*TorrentStatus{}
	for k, rv := range d {
		v, ok := rv.(rencode.Dictionary)
		if !ok {
			return nil, ErrInvalidDictionaryResponse
		}

		var ts TorrentStatus
		err = v.ToStruct(&ts, c.excludeV2tag)
		if err != nil {
			return nil, err
		}
		result[k] = &ts
	}

	return result, nil
}
