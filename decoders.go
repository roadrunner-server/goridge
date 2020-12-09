package goridge

import (
	"bytes"
	"encoding/gob"

	"github.com/spiral/errors"
	"github.com/vmihailenco/msgpack"
)

func decodeJson(out interface{}, frame *Frame) error {
	const op = errors.Op("codec: decode json")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	return json.Unmarshal(frame.Payload()[opts[1]:], out)
}

func decodeGob(out interface{}, frame *Frame) error {
	const op = errors.Op("codec: decode json")
	buf := new(bytes.Buffer)
	dec := gob.NewDecoder(buf)
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	buf.Write(payload)

	return dec.Decode(out)
}

func decodeRaw(out interface{}, frame *Frame) error {
	const op = errors.Op("codec: decode raw")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]

	if raw, ok := out.(*[]byte); ok {
		*raw = append(*raw, payload...)
		return nil
	}

	return nil
}

func decodeMsgPack(out interface{}, frame *Frame) error {
	const op = errors.Op("codec: decode msgpack")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	return msgpack.Unmarshal(payload, out)
}
