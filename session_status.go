package delugeclient

import (
	"github.com/gdm85/go-rencode"
)

// SessionStatus contains basic session status and statistics.
type SessionStatus struct {
	HasIncomingConnections bool
	UploadRate             float32
	DownloadRate           float32
	PayloadUploadRate      float32
	PayloadDownloadRate    float32
	TotalDownload          int64
	TotalUpload            int64
	NumPeers               int16
	DhtNodes               int16
}

// sessionStatusKeys is a slice with specific session status and statistics.
var sessionStatusKeys = rencode.NewList(
	"has_incoming_connections",
	"upload_rate",
	"download_rate",
	"payload_upload_rate",
	"payload_download_rate",
	"total_download",
	"total_upload",
	"num_peers",
	"dht_nodes",
)

// GetSessionStatus retrieves session status and statistics.
func (c *Client) GetSessionStatus() (*SessionStatus, error) {
	var args rencode.List
	args.Add(sessionStatusKeys)

	rd, err := c.rpcWithDictionaryResult("core.get_session_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}

	if c.settings.Logger != nil {
		c.settings.Logger.Printf("session status: %+s", rd)
	}
	var data SessionStatus
	err = rd.ToStruct(&data, c.excludeV2tag)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
