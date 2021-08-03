package rpc

import (
	"bytes"
	"encoding/gob"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/vmihailenco/msgpack"
	"google.golang.org/protobuf/proto"
)

func decodeJSON(out interface{}, frame *frame.Frame) error {
	const op = errors.Op("goridge_decode_json")
	opts := frame.ReadOptions(frame.Header())
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
	const op = errors.Op("goridge_decode_gob")
	opts := frame.ReadOptions(frame.Header())
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
	const op = errors.Op("goridge_decode_proto")
	opts := frame.ReadOptions(frame.Header())
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	// check if the out message is a correct proto.Message
	// instead send an error
	if pOut, ok := out.(proto.Message); ok {
		err := proto.Unmarshal(payload, pOut)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	return errors.E(op, errors.Str("message type is not a proto"))
}

func decodeRaw(out interface{}, frame *frame.Frame) error {
	const op = errors.Op("goridge_decode_raw")
	opts := frame.ReadOptions(frame.Header())
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
	const op = errors.Op("goridge_decodemsgpack")
	opts := frame.ReadOptions(frame.Header())
	if len(opts) != 2 {
		return errors.E(op, errors.Str("should be 2 options. SEQ_ID and METHOD_LEN"))
	}
	payload := frame.Payload()[opts[1]:]
	if len(payload) == 0 {
		return nil
	}

	return msgpack.Unmarshal(payload, out)
}
