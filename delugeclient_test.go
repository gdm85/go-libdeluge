// go-libdeluge v0.5.3 - a native deluge RPC client library
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
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"testing"
)

func TestZlibEOF(t *testing.T) {
	t.Parallel()

	b := bytes.NewReader([]byte{120, 156, 59, 204, 200, 200, 5, 0, 3, 31, 0, 208})

	zr, err := zlib.NewReader(b)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte{0}
	var decomp []byte
	for {
		n, err := zr.Read(data)
		if n == 1 {
			decomp = append(decomp, data[0])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
	}
	const expected = "decomp = [195 1 1 10]\n"
	s := fmt.Sprintln("decomp =", decomp)
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}
