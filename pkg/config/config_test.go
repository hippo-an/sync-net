package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// 테스트 실행
	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	require.Equal(t, "overwrite", config.Consistency.OnConflict)
	require.Equal(t, "/opt/sync-net/", config.Watcher.Path)
	require.Equal(t, 9999, config.Discovery.BroadcastPort)
	require.Equal(t, 9000, config.Discovery.TcpPort)
	require.Equal(t, 1*time.Minute, config.Discovery.BroadcastInterval)
	require.Equal(t, 4096, config.Discovery.BufferSize)
	require.Equal(t, 32768, config.Transfer.BufferSize)

	// 환경 변수 테스트
	os.Setenv("CONSISTENCY_ONCONFLICT", "backupAndOverwrite")
	os.Setenv("WATCHER_PATH", "/opt/lib/sync-net")
	os.Setenv("DISCOVERY_BROADCASTPORT", "8888")
	os.Setenv("DISCOVERY_TCPPORT", "8000")
	os.Setenv("DISCOVERY_BROADCASTINTERVAL", "30s")
	os.Setenv("DISCOVERY_BUFFERSIZE", "1024")
	os.Setenv("TRANSFER_BUFFERSIZE", "1024")

	config, err = NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	require.Equal(t, "backupAndOverwrite", config.Consistency.OnConflict)
	require.Equal(t, "/opt/lib/sync-net", config.Watcher.Path)
	require.Equal(t, 8888, config.Discovery.BroadcastPort)
	require.Equal(t, 8000, config.Discovery.TcpPort)
	require.Equal(t, 30*time.Second, config.Discovery.BroadcastInterval)
	require.Equal(t, 1024, config.Discovery.BufferSize)
	require.Equal(t, 1024, config.Transfer.BufferSize)

	// 환경 변수 초기화
	os.Unsetenv("CONSISTENCY_ONCONFLICT")
	os.Unsetenv("WATCHER_PATH")
	os.Unsetenv("DISCOVERY_BROADCASTPORT")
	os.Unsetenv("DISCOVERY_TCPPORT")
	os.Unsetenv("DISCOVERY_BROADCASTINTERVAL")
	os.Unsetenv("DISCOVERY_BUFFERSIZE")
	os.Unsetenv("TRANSFER_BUFFERSIZE")

}
