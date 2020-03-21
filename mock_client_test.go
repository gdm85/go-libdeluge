package delugeclient

import (
	"bytes"
	"encoding/hex"
)

// buffer is just here to make bytes.Buffer an io.ReadWriteCloser.
// Read about embedding to see how this works.
type buffer struct {
	bytes.Buffer
}

// Add a Close method to our buffer so that we satisfy io.ReadWriteCloser.
func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

func newMockClient(serial int64, payload string) DelugeClient {
	b, err := hex.DecodeString(payload)
	if err != nil {
		panic(err)
	}

	var c Client
	c.serial = serial
	c.safeConn = &buffer{
		Buffer: *bytes.NewBuffer(b),
	}

	return &c
}

func newMockClientV2(serial int64, payload string) DelugeClientV2 {
	b, err := hex.DecodeString(payload)
	if err != nil {
		panic(err)
	}

	var c ClientV2
	c.serial = serial
	c.safeConn = &buffer{
		Buffer: *bytes.NewBuffer(b),
	}

	return &c
}
