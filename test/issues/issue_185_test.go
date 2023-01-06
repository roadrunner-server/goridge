package issues

import (
	"fmt"
	"net"
	"net/rpc"
	"os/exec"
	"testing"
	"time"

	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type App struct{}

func (s *App) Hi(name string, r *string) error {
	*r = fmt.Sprintf("Hello, %s!", name)
	return nil
}

func TestRPC_Issue185(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:6001")
	require.NoError(t, err)

	err = rpc.Register(new(App))
	require.NoError(t, err)

	stopCh := make(chan struct{}, 1)

	go func() {
		time.Sleep(time.Second * 2)
		out, err := exec.Command("php", "../php_test_files/issue_185.php").Output()
		assert.NoError(t, err)

		assert.Equal(t, out, []byte("Hello, Antony!"))

		stopCh <- struct{}{}
	}()

	go func() {
		for range stopCh {
			_ = ln.Close()
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

		go rpc.ServeCodec(goridgeRpc.NewCodec(conn))
	}
}
