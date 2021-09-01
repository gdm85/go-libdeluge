// go-libdeluge v0.5.5 - a native deluge RPC client library
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

// Account is a user account inside the auth file.
type Account struct {
	Username  string
	Password  string
	AuthLevel AuthLevel
}

func (a *Account) fromDictionary(dict rencode.Dictionary) error {
	values, err := dict.Zip()
	if err != nil {
		return err
	}
	if len(values) < 3 {
		return ErrInvalidReturnValue
	}

	a.Username = string(values["username"].([]byte))
	a.Password = string(values["password"].([]byte))
	a.AuthLevel = AuthLevel(values["authlevel"].([]byte))

	return nil
}

func (a Account) toList() rencode.List {
	var list rencode.List
	list.Add(a.Username)
	list.Add(a.Password)
	list.Add(string(a.AuthLevel))
	return list
}
