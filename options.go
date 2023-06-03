// go-libdeluge v0.5.6 - a native deluge RPC client library
// Copyright (C) 2015~2023 gdm85 - https://github.com/gdm85/go-libdeluge/
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
	"reflect"

	"github.com/gdm85/go-rencode"
)

// Options used when adding a torrent magnet/URL.
// Valid options for v2: https://github.com/deluge-torrent/deluge/blob/deluge-2.0.3/deluge/core/torrent.py#L167-L183
// Valid options for v1: https://github.com/deluge-torrent/deluge/blob/1.3-stable/deluge/core/torrent.py#L83-L96
type Options struct {
	MaxConnections            *int
	MaxUploadSlots            *int
	MaxUploadSpeed            *int
	MaxDownloadSpeed          *int
	PrioritizeFirstLastPieces *bool
	PreAllocateStorage        *bool   // v2-only but automatically converted to compact_allocation for v1
	DownloadLocation          *string // works for both v1 and v2 when sending options
	AutoManaged               *bool
	StopAtRatio               *bool
	StopRatio                 *float32
	RemoveAtRatio             *float32
	MoveCompleted             *bool
	MoveCompletedPath         *string
	AddPaused                 *bool

	// V2 defines v2-only options
	V2 V2Options
}

type V2Options struct {
	SequentialDownload *bool
	Shared             *bool
	SuperSeeding       *bool
}

func (o *Options) toDictionary(v2daemon bool) rencode.Dictionary {
	var dict rencode.Dictionary
	if o == nil {
		return dict
	}

	v := reflect.ValueOf(*o)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Struct {
			// there is a single struct fields, V2, which is conditionally parsed after this loop
			continue
		}
		if f.IsNil() {
			continue
		}

		name := rencode.ToSnakeCase(t.Field(i).Name)
		if !v2daemon && name == "pre_allocate_storage" {
			name = "compact_allocation"
		}

		dict.Add(name, reflect.Indirect(f).Interface())
	}

	if !v2daemon {
		return dict
	}

	// add the v2-only fields
	v = reflect.ValueOf(o.V2)
	t = v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.IsNil() {
			continue
		}

		name := rencode.ToSnakeCase(t.Field(i).Name)
		dict.Add(name, reflect.Indirect(f).Interface())
	}

	return dict
}
