package delugeclient

import (
	"github.com/gdm85/go-rencode"
)

// SessionStats Basic session stats
type SessionStats struct {
	PayloadDownloadRate    int64
	PayloadUploadRate      int64
	DownloadRate           int64
	UploadRate             int64
	TotalDownload          int64
	TotalUpload            int64
	DhtNodes               int16
	NumPeers               int16
	HasIncomingConnections bool
}

// sessStatsKeys default keys of session stats, see "deluge/core/core.py"
var sessStatsKeys = rencode.NewList(
	"payload_download_rate",
	"payload_upload_rate",
	"download_rate",
	"upload_rate",
	"total_download",
	"total_upload",
	"dht_nodes",
	"num_peers",
	"has_incoming_connections",
)

// SessionStats Gets session status values as SessionStats object
func (c *Client) SessionStats() (*SessionStats, error) {
	var args rencode.List
	args.Add(sessStatsKeys)

	rd, err := c.rpcWithDictionaryResult("core.get_session_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("session stats: %+s", rd)
	}
	var data SessionStats
	err = rd.ToStruct(&data, c.excludeV2tag)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
