package delugeclient

import (
	"github.com/gdm85/go-rencode"
)

// SessionStats ...
type SessionStats struct {
	PayloadDownloadRate int64
	PayloadUploadRate   int64
	DownloadRate        int64
	UploadRate          int64
	TotalDownload       int64 // net.recv_bytes
	TotalUpload         int64 // net.sent_bytes
	// FreeSpace           int64
	DhtNodes int16
	NumPeers int16
	// NumTorrents            int16
	// NumPausedTorrents      int16
	HasIncomingConnections bool
}

var sessStatsKeys = rencode.NewList(
	"payload_download_rate",
	"payload_upload_rate",
	"download_rate",
	"upload_rate",
	"total_download",
	"total_upload",
	// "free_space",
	"dht_nodes",
	"num_peers",
	// "num_torrents",
	// "num_paused_torrents",
	"has_incoming_connections",
)

// SessionStats ...
func (c *Client) SessionStats(keys ...string) (*SessionStats, error) {
	var args rencode.List
	if 0 == len(keys) {
		args.Add(sessStatsKeys)
	} else {
		args.Add(rencode.NewList(keys))
	}

	rd, err := c.rpcWithDictionaryResult("core.get_session_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("got stats: %+s", rd)
	}
	var data SessionStats
	err = rd.ToStruct(&data, c.excludeV2tag)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
