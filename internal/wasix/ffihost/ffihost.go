package ffihost

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero/api"
)

func ReadMicroDeadline(ctx context.Context, deadline int64) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, time.UnixMicro(deadline))
}

func NewFile(m api.Memory, root fs.FS, fd int64, offset uint32, l uint32) (_ fs.File, err error) {
	var (
		name string
	)
	if name, err = ReadString(m, offset, l); err != nil {
		return nil, err
	}

	return root.Open(name)
}

func OpenFile(m api.Memory, root fs.FS, fd int64, offset uint32, l uint32) (_ *os.File, err error) {
	var (
		name string
	)

	if name, err = ReadString(m, offset, l); err != nil {
		return nil, err
	}

	return os.OpenFile(name, os.O_RDWR, 0600)
}

func ReadString(m api.Memory, offset uint32, len uint32) (string, error) {
	var (
		ok   bool
		data []byte
	)

	if data, ok = m.Read(offset, len); !ok {
		return "", errors.New("unable to read string")
	}

	return string(data), nil
}

func ReadArrayElement(m api.Memory, offset, len uint32) (data []byte, err error) {
	var (
		ok            bool
		eoffset, elen uint32
	)

	// log.Println("reading array element", offset, len)

	if eoffset, ok = m.ReadUint32Le(offset); !ok {
		return nil, errors.New("unable to read element offset")
	}

	if elen, ok = m.ReadUint32Le(offset + 4); !ok {
		return nil, errors.New("unable to read element byte length")
	}

	if data, ok = m.Read(eoffset, elen); !ok {
		return nil, errors.New("unable to read element bytes")
	}

	return data, nil
}

func ReadBytes(m api.Memory, offset uint32, len uint32) (data []byte, err error) {
	var (
		ok bool
	)

	if data, ok = m.Read(offset, len); !ok {
		return nil, errors.New("unable to read string")
	}

	return data, nil
}

func ReadJSON(m api.Memory, offset uint32, len uint32, v interface{}) (err error) {
	var (
		ok      bool
		encoded []byte
	)

	if encoded, ok = m.Read(offset, len); !ok {
		return fmt.Errorf("unable to read json encoded data from memory: %d, %d", offset, len)
	}

	if err = json.Unmarshal(encoded, &v); err != nil {
		return errors.Wrap(err, "unable to deserialize json")
	}

	return nil
}

func WriteBool(m api.Memory, offset uint32, b bool) (err error) {
	booltobyte := func(b bool) byte {
		if b {
			return 1
		}

		return 0
	}

	if !m.WriteByte(offset, booltobyte(b)) {
		return fmt.Errorf("unable to write boolean: %d", offset)
	}

	return nil
}

func WriteJSON(m api.Memory, dmax uint32, dptr, dlenptr uint32, d any) (err error) {
	encoded, err := json.Marshal(d)
	if err != nil {
		return err
	}

	if len(encoded) > int(dmax) {
		return fmt.Errorf("unable to write json buffer exceeded %d > %d", len(encoded), dmax)
	}

	if !m.WriteString(dptr, string(encoded)) {
		return fmt.Errorf("unable to write json: %d", dptr)
	}

	if !m.WriteUint32Le(dlenptr, uint32(len(encoded))) {
		return fmt.Errorf("unable to write binary json length: %d", dlenptr)
	}

	return nil
}
