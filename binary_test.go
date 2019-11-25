package goridge

import (
	"github.com/stretchr/testify/assert"
	"net/rpc"
	"testing"
)

func TestPackUnpack(t *testing.T) {
	var (
		req = &rpc.Request{
			ServiceMethod: "test.Process",
			Seq:           199,
		}
		res = &rpc.Response{}
	)

	data := make([]byte, len(req.ServiceMethod)+Uint64Size)
	pack(req.ServiceMethod, req.Seq, data)

	assert.Len(t, data, len(req.ServiceMethod)+Uint64Size)
	assert.NoError(t, unpack(data, &res.ServiceMethod, &res.Seq))

	assert.Equal(t, res.ServiceMethod, req.ServiceMethod)
	assert.Equal(t, res.Seq, req.Seq)

	assert.Equal(t, "test.Process", res.ServiceMethod)
	assert.Equal(t, uint64(199), res.Seq)
}


