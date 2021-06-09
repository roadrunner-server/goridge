package rpc

import (
	"bytes"
	"encoding/gob"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/vmihailenco/msgpack"
	"google.golang.org/protobuf/proto"
)

func decodeJSON(out interface{}, frame *frame.Frame) error {
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

func decodeGob(out interface{}, frame *frame.Frame) error {
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

func decodeProto(out interface{}, frame *frame.Frame) error {
	const op = errors.Op("client: decode PROTO")
	opts := frame.ReadOptions()
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	err := proto.Unmarshal(payload, out.(proto.Message))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func decodeRaw(out interface{}, frame *frame.Frame) error {
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

func decodeMsgPack(out interface{}, frame *frame.Frame) error {
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
