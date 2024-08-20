package transfer

import (
	"fmt"
	"github.com/hippo-an/sync-net/pkg/discovery"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	w := &watcher.Watcher{}
	s := &discovery.Server{}
	client := NewClient(w, s)

	require.Equal(t, client.w, w)
	require.Equal(t, client.s, s)
}

func TestHandleEvents(t *testing.T) {
	w := &watcher.Watcher{
		CreateEventChan: make(chan *watcher.Event),
		ModifyEventChan: make(chan *watcher.Event),
		DeleteEventChan: make(chan *watcher.Event),
		ErrorChan:       make(chan error),
		StopChan:        make(chan struct{}),
	}
	s := &discovery.Server{
		ServerInfos: map[string]*discovery.ServerInfo{
			"127.0.0.1": {Ip: "127.0.0.1", Port: "8080", Self: false},
		},
	}
	client := NewClient(w, s)

	go client.HandleEvents()

	w.CreateEventChan <- &watcher.Event{EventType: watcher.Create, FullPath: "/test/file.txt", Name: "file.txt"}
	time.Sleep(time.Millisecond * 100)

	w.ModifyEventChan <- &watcher.Event{EventType: watcher.Modify, FullPath: "/test/file.txt", Name: "file.txt"}
	time.Sleep(time.Millisecond * 100)

	w.DeleteEventChan <- &watcher.Event{EventType: watcher.Delete, FullPath: "/test/file.txt", Name: "file.txt"}
	time.Sleep(time.Millisecond * 100)

	w.ErrorChan <- fmt.Errorf("test error")
	time.Sleep(time.Millisecond * 100)

	close(w.StopChan)
}

func TestHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		require.NoError(t, err)

		require.Equal(t, string(buffer[:n]), "1:/test/file.txt")
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	err = handshake(conn, "1:/test/file.txt")
	require.NoError(t, err)
}

func TestFileTransfer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file-transfer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("This is a test file content")
	err = os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		receivedContent, err := io.ReadAll(conn)
		require.NoError(t, err)

		require.Equal(t, receivedContent, testContent)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	err = fileTransfer(conn, testFile)
	require.NoError(t, err)
}
