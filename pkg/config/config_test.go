package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// 테스트 실행
	config, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	require.Equal(t, "overwrite", config.Consistency.OnConflict)
	require.Equal(t, "/opt/sync-net/", config.Watcher.Path)

	// 환경 변수 테스트
	os.Setenv("CONSISTENCY_ONCONFLICT", "backupAndOverwrite")
	os.Setenv("WATCHER_PATH", "/opt/lib/sync-net")

	config, err = NewConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	require.Equal(t, "backupAndOverwrite", config.Consistency.OnConflict)
	require.Equal(t, "/opt/lib/sync-net", config.Watcher.Path)

	// 환경 변수 초기화
	os.Unsetenv("CONSISTENCY_ONCONFLICT")
	os.Unsetenv("WATCHER_PATH")
}
