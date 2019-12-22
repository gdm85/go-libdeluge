// go-libdeluge v0.3.1 - a native deluge RPC client library
// Copyright (C) 2015~2019 gdm85 - https://github.com/gdm85/go-libdeluge/
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
	"fmt"

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
func (c *Client) AddTorrentMagnet(magnetURI string, options *Options) (string, error) {
	var args rencode.List
	args.Add(magnetURI, options.toDictionary(c.settings.V2Daemon))

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
func (c *Client) AddTorrentURL(url string, options *Options) (string, error) {
	var args rencode.List
	args.Add(url, options.toDictionary(c.settings.V2Daemon))

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

func (c *Client) TorrentStatus(id string) (*TorrentStatus, error) {
	var args rencode.List
	args.Add(id)
	args.Add(STATUS_KEYS)

	resp, err := c.rpc("core.get_torrent_status", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	return decodeTorrentStatusResponse(resp)
}

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

func decodeTorrentStatusResponse(resp *DelugeResponse) (*TorrentStatus, error) {
	values := resp.returnValue.Values()
	if len(values) != 1 {
		return nil, ErrInvalidReturnValue
	}
	rd, ok := values[0].(rencode.Dictionary)
	if !ok {
		return nil, ErrInvalidListResult
	}

	var ts TorrentStatus
	_ = rd.ToStruct(&ts)

	return &ts, nil
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

// TorrentError is a tuple of a torrent and an error message, returned by
// methods that manipulate many torrents at once.
type TorrentError struct {
	// ID is the hash of the torrent that experienced an error
	ID      string
	Message string
}

func (t TorrentError) Error() string {
	return fmt.Sprintf("<%s>: '%s'", t.ID, t.Message)
}

// RemoveTorrents tries to remove multiple torrents at once.
// If `rmFiles` is set it also tries to delete all downloaded data for the
// specified torrents.
// If errors were encountered, the returned list will be a list of
// TorrentErrors.
// On success, an empty list of errors is returned.
//
// The user should not rely on files being removed or torrents being
// removed from the session, just because no errors have been returned,
// as returned errors will primarily indicate that some of the supplied
// torrent hashes were invalid.
func (c *Client) RemoveTorrents(ids []string, rmFiles bool) ([]TorrentError, error) {
	var args rencode.List
	args.Add(sliceToRencodeList(ids), rmFiles)

	resp, err := c.rpc("core.remove_torrents", args, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	vals := resp.returnValue.Values()
	if len(vals) != 1 {
		return nil, ErrInvalidReturnValue
	}
	failedList := vals[0].(rencode.List)

	var torrentErrors []TorrentError

	// Iterate through the list of errors that have occured, and
	// convert each of them into a more typesafe format.
	for _, e := range failedList.Values() {
		failedEntry, ok := e.(rencode.List)
		if !ok {
			// Unexpected response from the API
			return torrentErrors, ErrInvalidReturnValue
		}

		failedTuple := failedEntry.Values()
		if len(failedTuple) != 2 {
			// return here, as we don't know how to parse the returned
			// error structure
			return torrentErrors, ErrInvalidReturnValue
		}

		torrentError := TorrentError{
			ID:      string(failedTuple[0].([]byte)),
			Message: string(failedTuple[1].([]byte)),
		}

		torrentErrors = append(torrentErrors, torrentError)
	}

	return torrentErrors, nil
}

// RemoveTorrent removes a single torrent, returning true if successful.
// If `rmFiles` is set it also tries to delete all downloaded data for the
// specified torrent.
func (c *Client) RemoveTorrent(id string, rmFiles bool) (bool, error) {
	var args rencode.List
	args.Add(id, rmFiles)

	resp, err := c.rpc("core.remove_torrent", args, rencode.Dictionary{})
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, resp.RPCError
	}

	vals := resp.returnValue.Values()
	if len(vals) != 1 {
		return false, ErrInvalidReturnValue
	}
	success := vals[0]

	return success.(bool), nil
}

// PauseTorrents pauses a group of torrents with the given IDs.
func (c *Client) PauseTorrents(ids []string) error {
	var args rencode.List
	args.Add(sliceToRencodeList(ids))

	resp, err := c.rpc("core.pause_torrents", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return err
}

// PauseTorrent pauses a single torrent with the given ID.
func (c *Client) PauseTorrent(id string) error {
	var args rencode.List
	args.Add(id)

	resp, err := c.rpc("core.pause_torrent", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return err
}

// ResumeTorrents unpauses a group of torrents with the given IDs.
func (c *Client) ResumeTorrents(ids []string) error {
	var args rencode.List
	args.Add(sliceToRencodeList(ids))

	resp, err := c.rpc("core.resume_torrents", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return err
}

// ResumeTorrent unpauses a single torrent with the given ID.
func (c *Client) ResumeTorrent(id string) error {
	var args rencode.List
	args.Add(id)

	resp, err := c.rpc("core.resume_torrent", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return err
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

// SetTorrentOptions updates options for the torrent with the given hash.
func (c *Client) SetTorrentOptions(id string, options *Options) error {
	var args rencode.List
	args.Add(id, options.toDictionary(c.settings.V2Daemon))

	resp, err := c.rpc("core.set_torrent_options", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return nil
}

// SetTorrentTracker sets the primary tracker for the torrent with the
// given hash to be `trackerURL`.
func (c *Client) SetTorrentTracker(id, trackerURL string) error {
	var tracker rencode.Dictionary
	tracker.Add("url", trackerURL)
	tracker.Add("tier", 0)

	var trackers rencode.List
	trackers.Add(tracker)

	var args rencode.List
	args.Add(id, trackers)

	resp, err := c.rpc("core.set_torrent_trackers", args, rencode.Dictionary{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.RPCError
	}

	return nil
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

// KnownAccounts returns all known accounts, including password and
// permission levels.
func (c *Client) KnownAccounts() ([]Account, error) {
	if !c.settings.V2Daemon {
		return nil, ErrUnsupportedV1
	}

	resp, err := c.rpc("core.get_known_accounts", rencode.List{}, rencode.Dictionary{})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.RPCError
	}

	var users rencode.List
	err = resp.returnValue.Scan(&users)
	if err != nil {
		return nil, err
	}

	// users is now a list of dictionaries, each containing
	// three []byte attributes: username, password and auth level
	var accounts []Account
	for _, u := range users.Values() {
		dict, ok := u.(rencode.Dictionary)
		if !ok {
			return nil, ErrInvalidListResult
		}

		var a Account
		err := a.fromDictionary(dict)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

// CreateAccount creates a new Deluge user with the supplied username,
// password and permission level. The authenticated user must have an
// authLevel of ADMIN to succeed.
func (c *Client) CreateAccount(account Account) (bool, error) {
	if !c.settings.V2Daemon {
		return false, ErrUnsupportedV1
	}

	resp, err := c.rpc("core.create_account", account.toList(), rencode.Dictionary{})
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, resp.RPCError
	}

	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return false, ErrInvalidReturnValue
	}
	success := vals[0]

	return success.(bool), nil
}

// UpdateAccount sets a new password and permission level for a account.
// The authenticated user must have an authLevel of ADMIN to succeed.
func (c *Client) UpdateAccount(account Account) (bool, error) {
	if !c.settings.V2Daemon {
		return false, ErrUnsupportedV1
	}

	resp, err := c.rpc("core.update_account", account.toList(), rencode.Dictionary{})
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, resp.RPCError
	}

	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return false, ErrInvalidReturnValue
	}
	success := vals[0]

	return success.(bool), nil
}

// RemoveAccount will delete an existing username.
// The authenticated user must have an authLevel of ADMIN to succeed.
func (c *Client) RemoveAccount(username string) (bool, error) {
	if !c.settings.V2Daemon {
		return false, ErrUnsupportedV1
	}

	var args rencode.List
	args.Add(username)

	resp, err := c.rpc("core.remove_account", args, rencode.Dictionary{})
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, resp.RPCError
	}

	vals := resp.returnValue.Values()
	if len(vals) == 0 {
		return false, ErrInvalidReturnValue
	}
	success := vals[0]

	return success.(bool), nil
}
