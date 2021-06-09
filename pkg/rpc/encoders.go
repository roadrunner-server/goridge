package rpc

import (
	"encoding/gob"
	"io"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/vmihailenco/msgpack"
	"google.golang.org/protobuf/proto"
)

func encodeJSON(out io.Writer, data interface{}) error {
	const op = errors.Op("codec: encode json")

	res, err := json.Marshal(data)
	if err != nil {
		return errors.E(op, err)
	}
	_, err = out.Write(res)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func encodeGob(out io.Writer, data interface{}) error {
	const op = errors.Op("codec: encode GOB")

	dec := gob.NewEncoder(out)
	err := dec.Encode(data)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func encodeProto(out io.Writer, data interface{}) error {
	const op = errors.Op("codec: encode PROTO")

	d, err := proto.Marshal(data.(proto.Message))
	if err != nil {
		return errors.E(op, err)
	}

	_, err = out.Write(d)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func encodeRaw(out io.Writer, data interface{}) error {
	const op = errors.Op("codec: encode raw")
	switch data := data.(type) {
	case []byte:
		_, err := out.Write(data)
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	case *[]byte:
		_, err := out.Write(*data)
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	default:
		return errors.E(op, errors.Str("unknown Raw payload type"))
	}
}

func encodeMsgPack(out io.Writer, data interface{}) error {
	const op = errors.Op("codec: encode msgpack")
	b, err := msgpack.Marshal(data)
	if err != nil {
		return errors.E(op, err)
	}
	_, err = out.Write(b)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
