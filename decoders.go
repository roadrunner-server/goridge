package goridge

import (
	"bytes"
	"encoding/gob"

	"github.com/spiral/errors"
	"github.com/vmihailenco/msgpack"
)

func decodeJSON(out interface{}, frame *Frame) error {
	const op = errors.Op("client: decode json")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, out)
}

func decodeGob(out interface{}, frame *Frame) error {
	const op = errors.Op("client: decode GOB")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	buf := new(bytes.Buffer)
	dec := gob.NewDecoder(buf)
	buf.Write(payload)

	return dec.Decode(out)
}

func decodeRaw(out interface{}, frame *Frame) error {
	const op = errors.Op("client: decode raw")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	if raw, ok := out.(*[]byte); ok {
		*raw = append(*raw, payload...)
		return nil
	}

	return nil
}

func decodeMsgPack(out interface{}, frame *Frame) error {
	const op = errors.Op("client: decode msgpack")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	return msgpack.Unmarshal(payload, out)
}
