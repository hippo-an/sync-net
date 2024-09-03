package transfer

import (
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const (
	backupAndCreate = "backupAndCreate"
	overwrite       = "overwrite"
)

func getConfig(onConflict string) (*config.Config, error) {
	conf, err := config.NewConfig()
	if onConflict != "" {
		conf.Transfer.Consistency.OnConflict = onConflict
	}
	return conf, err
}

func TestHandleCreateEventWithOverwrite(t *testing.T) {
	conf, err := getConfig(overwrite)

	if err != nil {
		require.NoError(t, err)
	}
	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "TestHandleCreateEventWithOverwrite.txt")
	testContent := []byte("This is a test file")

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()

		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		err = s.handleCreateEvent(conn, testFilePath)
		require.NoError(t, err)

		data, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, testContent, data)
	}(&wg)

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)

	_, err = conn.Write(testContent)
	require.NoError(t, err)
	conn.Close()
	wg.Wait()
}

func TestHandleCreateEventWithBackup(t *testing.T) {
	conf, err := getConfig(backupAndCreate)

	if err != nil {
		require.NoError(t, err)
	}

	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "TestHandleCreateEventWithOverwrite.txt")
	testContent := []byte("This is a test file")

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()

		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		err = s.handleCreateEvent(conn, testFilePath)
		require.NoError(t, err)

		data, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, testContent, data)
	}(&wg)

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)

	_, err = conn.Write(testContent)
	require.NoError(t, err)
	conn.Close()
	wg.Wait()
}

func TestHandleModifyEventWithOverwrite(t *testing.T) {
	conf, err := getConfig(overwrite)

	if err != nil {
		require.NoError(t, err)
	}
	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "TestHandleModifyEvent.txt")
	testContent := []byte("This is a test file")
	err = os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()

		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		err = s.handleModifyEvent(conn, testFilePath)
		require.NoError(t, err)

		data, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, []byte("Modified content"), data)
	}(&wg)

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)

	_, err = conn.Write([]byte("Modified content"))
	require.NoError(t, err)
	conn.Close()
	wg.Wait()
}

func TestHandleModifyEventWithBackup(t *testing.T) {
	conf, err := getConfig(backupAndCreate)

	if err != nil {
		require.NoError(t, err)
	}
	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalText := "This is a test file"
	modifiedText := "Modified content"
	testFilePath := filepath.Join(tempDir, "TestHandleModifyEvent.txt")
	testContent := []byte(originalText)
	err = os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()

		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		err = s.handleModifyEvent(conn, testFilePath)
		require.NoError(t, err)

		data, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, []byte(modifiedText), data)

		data, err = os.ReadFile(testFilePath + ".backup")
		require.NoError(t, err)
		require.Equal(t, []byte(originalText), data)
	}(&wg)

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)

	_, err = conn.Write([]byte(modifiedText))
	require.NoError(t, err)
	conn.Close()
	wg.Wait()
}

func TestHandleDeleteEventWithOverwrite(t *testing.T) {
	conf, err := getConfig(overwrite)

	if err != nil {
		require.NoError(t, err)
	}
	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "TestHandleDeleteEvent.txt")
	testContent := []byte("This is a test file")
	err = os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	err = s.handleDeleteEvent(testFilePath)
	require.NoError(t, err)

	_, err = os.Stat(testFilePath)
	require.True(t, os.IsNotExist(err))
}

func TestHandleDeleteEventWithBackup(t *testing.T) {
	conf, err := getConfig(backupAndCreate)

	if err != nil {
		require.NoError(t, err)
	}
	s := NewServer(conf)
	tempDir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFilePath := filepath.Join(tempDir, "TestHandleDeleteEvent.txt")
	originalText := "This is a test file"
	testContent := []byte(originalText)
	err = os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	err = s.handleDeleteEvent(testFilePath)
	require.NoError(t, err)

	_, err = os.Stat(testFilePath)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(testFilePath + ".backup")
	require.False(t, os.IsNotExist(err))
}

func TestParseEventInfo(t *testing.T) {
	eventType, filePath, err := parseEventInfo("1:/test/file.txt")
	require.Equal(t, watcher.EventType(1), eventType)
	require.Equal(t, "/test/file.txt", filePath)
	assert.NoError(t, err)

	eventType, filePath, err = parseEventInfo("invalid")
	require.Equal(t, watcher.EventType(0), eventType)
	require.Equal(t, "", filePath)
	require.Error(t, err)
}
