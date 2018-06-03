package goridge

import (
	"testing"
	"net/rpc"
	"github.com/stretchr/testify/assert"
)

func TestPackUnpack(t *testing.T) {
	var (
		req = &rpc.Request{
			ServiceMethod: "test.Process",
			Seq:           199,
		}
		res = &rpc.Response{}
	)

	data := pack(req.ServiceMethod, req.Seq)
	assert.Len(t, data, len(req.ServiceMethod)+8)
	assert.NoError(t, unpack(data, &res.ServiceMethod, &res.Seq))

	assert.Equal(t, res.ServiceMethod, req.ServiceMethod)
	assert.Equal(t, res.Seq, req.Seq)

	assert.Equal(t, "test.Process", res.ServiceMethod)
	assert.Equal(t, uint64(199), res.Seq)
}
