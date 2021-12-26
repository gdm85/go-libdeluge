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
	"testing"
)

func TestConnect(t *testing.T) {
	t.Parallel()

	c := newMockClient(0, "789C3BCCC8C80500031F00D0")

	err := c.DaemonLogin()
	if err != nil {
		t.Error(err)
	}
}

func TestDaemonVersion(t *testing.T) {
	t.Parallel()

	c := newMockClient(1, "789C3BCCC8B4C848CF40CF58D748D7C8C0D0D2C0CCD0C8D0DCC45CB734A934AFA4D4D042CFC044CF1000B5C20978")

	ver, err := c.DaemonVersion()
	if err != nil {
		t.Fatal(err)
	}
	if ver != "2.0.3-2-201906121747-ubuntu18.04.1" {
		t.Error("version mismatch")
	}
}

func TestMethodsList(t *testing.T) {
	t.Parallel()

	c := newMockClient(2, "789C85944B6EDB30108651F43EB9400F43D0E458264C912C67E826DE74DD4563C4719CA441FA408B5EA417EB48A45E96D0EEC4EF1F71DEFCF3E6EDBB07E5235C49AD05F918C191581B0B2F8B5448BC71EABCA8E1E30CD7B27240A7194FD1DEB55079E740910800F13E930892D88F523E399AB0F27766DAA05C7140C1A6CAB863CBC08D504E6BEDA302C1FF3BC7F729384EA8DA80DABEB6A8021232D14658D881450E3D04E32AFC32A83B69ECC8013EF792F275B040C63B11246DF03052DCDA540F1767B19336C17991E2538F733EBAF3771A846B82E8A415260C907BC0545004B81F209F0406A9E0B1675BE73FB8AEC2F8D2736B565D83761091B3398D3424E0EC7CA4630F9B5C059A3DDC0E28FAEB9B2131046CEE1148DCC0C7459C70E09DFBCC9F2E3916E15316AC5F654F06FB1B834C083A0F57ED779C3BFF2BABD2F756ED6CC768325A1384793C42843590DA9491163590D49264F6F43E4102B1F244BE3E8C88E642DF8ECEE4C3E7D13185EC30421BEA64E60B2B519C1618DE15E8640D79038F13E2ADEEB62A022AE9BA41EA59AAFB6A4CD885D7312BB38DFDD4BEF6E7AE7BBC469EEBE4B8ECFAFB4C6D9F91108D8F860CE0DCA096D7A2BC0B1C17FE5C34682A6BBDD43CDB00FAC7A24D0AD9C27AC27F5B34777C9B5B34C52EAB0DFAD77FF4761B9E66463EB459FC9E095D09F64DEF22F28A49A47935BAD9E02FC9377D9D19F080875E7E59965BED79A651946ACB9B9E9F27029CEC799E8814F4E83D3E16D6962D4FD3594B0ED15D358F2727B4E74A28696D87ABBC2C1BAFDBBB8F235C1E994341B849D4B4F5E35F97E8BAB4")

	_, err := c.MethodsList()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFreeSpace(t *testing.T) {
	t.Parallel()

	c := newMockClient(3, "789C3BCCC8E206000361010F")

	_, err := c.GetFreeSpace("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAvailablePlugins(t *testing.T) {
	t.Parallel()

	c := newMockClient(4, "789C3BCCC87AAA353C352934B32D243F3D3D27B535B824B1A4B83338392335A53427B5A8D72FBF24332D3339B124333FAFB8D527312935A7D3B5A2A42831B924BFA8DDB52235B9B424B5D329273F393B27B3B8A4DDB1B424DF312505007A9F24DD")

	_, err := c.GetAvailablePlugins()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetEnabledPlugins(t *testing.T) {
	t.Parallel()

	c := newMockClient(5, "789C3BCCC876000003DF018B")

	_, err := c.GetEnabledPlugins()
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnablePlugin(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(7, "789C3BCCC8E10C0003660110")

	err := c.EnablePlugin("Label")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDisablePlugin(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(8, "789C3BCCC8E90C0003680111")

	err := c.DisablePlugin("Label")
	if err != nil {
		t.Fatal(err)
	}
}

func TestKnownAccounts(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(6, "789C3BCCC87E30ABA3B438B5282F3137B53B273F39312739273335AFA4A320B1B8B83CBF28A52D2535A7343DB533B1B4242327B52C35A7D5D1C5D7D3AF17CE8FCFCC2BE1020076541E37")

	_, err := c.KnownAccounts()
	if err != nil {
		t.Fatal(err)
	}
}

const (
	testMagnetURI = `magnet:?xt=urn:btih:c1939ca413b9afcc34ea0cf3c128574e93ff6cb0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce`
)

func TestAddTorrentMagnet(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(7, "789C3BCCC8B122D9D0D2D83239D1C4D038C932312D39D9D82435D12039CD38D9D0C8C2D4DC24D5D2382DCD2C39C9000028B20CF1")

	_, err := c.AddTorrentMagnet(testMagnetURI, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTorrentsStatus(t *testing.T) {
	t.Parallel()

	c := newMockClient(8, "789C8D914B4EC3301086CD923D0740624F5B521E959080AAEA05D86339F624B548ECC81E97C70624A880A62DA208CEC02D88C489B801CEA36CD8E0CD581ACDFCFFF7CFD7DA7AFCC13BBDA0C759B713843D16711E7481B57914F0CECEC1EE7E177A4114EDF1B07DF8901939660883296A6409B5F21AC86366746CC0DA3EF1EF45E80B95682668A23943A9D57B6BA4536819A71498D6A069DBBC5E9101183BCC194739068A32053231E55CFFF39690A590168D0C1D82A05C67126A95191AC6CFC1D091B63875A153E8B6B94E6716404815D78BA6CAA5D4CF70B0E4415A5A36BD75DFA24C0810C767CBAD8DB9824BA44C29ED1487A3E0FB399209504FAA8D442F58DC2B96C2BF436AA2115A01C9BD682495B4232FFCF69B4CC6AEAAEA393DAD455FF2552EDEFCA434608B49154D316FECD419170B4F992550E651313E958C2598ADBFD5109955D4340115E3882C56699552CECE4F1AD8CD5350F8EAB2BFA6EE0019692E542D1FDEFC00FE45D638")

	_, err := c.TorrentsStatus(StateUnspecified, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddTorrentFile(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(7, "789C3BCCC8B1C224D1C03825D1D8D2DCD420D9242925C5202D3529C9CC2CC522C9D0D822393535D1D4D2D2382D3111002A7C0D50")

	_, err := c.AddTorrentFile("ubuntu-14.04.6-desktop-amd64.iso.torrent", "base64 data", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPauseTorrents(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(8, "789C3BCCC8E90A00036A0113")

	err := c.PauseTorrents("some-torrent-hash")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveTorrent(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(9, "789C3BCCC8E50C00036A0112")

	_, err := c.RemoveTorrent("some-torrent-hash", false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResumeTorrents(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(9, "789C3BCCC8E50A00036C0114")

	err := c.ResumeTorrents("some-torrent-hash")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetListenPort(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(13, "789C3BCCC8E7C0C0F097030009BD0218")

	_, err := c.GetListenPort()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSessionStatus(t *testing.T) {
	t.Parallel()

	c := newMockClientV2(4, "789C3BCCC89A3FB520B132273F31253E25BF3C0FCC284A2C4975620082C930A9D20254895E4CB5DDE86AFA4AF24B1273E0A632F440F810650C9D291925F179F929A9C50C9D79A5B9F105A9A945C50C3332128BE333F392F37333F3D2E393F3F3F252934B32F3F38A19004F9C3E17")

	_, err := c.GetSessionStatus()
	if err != nil {
		t.Fatal(err)
	}
}
