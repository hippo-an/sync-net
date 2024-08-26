package watcher

import (
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testFileName = "testFile.txt"
)

func create(t *testing.T, w *Watcher) string {
	t.Helper()

	testFile := filepath.Join(w.BasePath, testFileName)
	_, err := os.Create(testFile)
	require.NoError(t, err)

	select {
	case e := <-w.CreateEventChan:
		require.Equal(t, testFileName, e.Name)
		require.Equal(t, w.BasePath, e.Path)
		require.Equal(t,
			w.BasePath+"/"+testFileName, e.FullPath)
		require.Equal(t, File, e.FileType)
		require.Equal(t, Create, e.EventType)
		require.WithinDuration(t, time.Now(), e.ModifiedAt, time.Second)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	return testFile
}

func createConf(t *testing.T) *config.Config {
	t.Helper()
	conf, err := config.NewConfig()
	require.NoError(t, err)
	conf.Watcher.Path = t.TempDir()

	return conf
}

func TestFileCreate(t *testing.T) {
	conf := createConf(t)
	w, err := NewWatcher(conf)
	require.NoError(t, err)

	defer w.TearDown()

	go StartWatch(w)

	create(t, w)
}

func TestFileModify(t *testing.T) {
	conf := createConf(t)

	w, err := NewWatcher(conf)
	require.NoError(t, err)

	defer w.TearDown()

	go StartWatch(w)

	testFile := create(t, w)

	err = os.WriteFile(testFile, []byte("data"), 0644)
	require.NoError(t, err)

	select {
	case e := <-w.ModifyEventChan:
		require.Equal(t, testFileName, e.Name)
		require.Equal(t, conf.Watcher.Path, e.Path)
		require.Equal(t, conf.Watcher.Path+"/"+testFileName, e.FullPath)
		require.Equal(t, File, e.FileType)
		require.Equal(t, Modify, e.EventType)
		require.WithinDuration(t, time.Now(), e.ModifiedAt, time.Second)

	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for modify event")
	}
}

func TestFileDelete(t *testing.T) {
	conf := createConf(t)

	w, err := NewWatcher(conf)
	require.NoError(t, err)

	defer w.TearDown()

	go StartWatch(w)
	testFile := create(t, w)

	err = os.Remove(testFile)
	require.NoError(t, err)

	select {
	case e := <-w.DeleteEventChan:
		require.Equal(t, testFileName, e.Name)
		require.Equal(t, conf.Watcher.Path, e.Path)
		require.Equal(t, conf.Watcher.Path+"/"+testFileName, e.FullPath)
		require.Equal(t, Deleted, e.FileType)
		require.Equal(t, Delete, e.EventType)
		require.WithinDuration(t, time.Now(), e.ModifiedAt, time.Second)

	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for modify event")
	}
}

func TestNestedFolderCreate(t *testing.T) {
	conf := createConf(t)

	w, err := NewWatcher(conf)
	require.NoError(t, err)
	defer w.TearDown()

	go StartWatch(w)

	nestedDir := filepath.Join(conf.Watcher.Path, "nested", "folder", "structure")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	<-w.CreateEventChan

	testFile := filepath.Join(nestedDir, testFileName)
	_, err = os.Create(testFile)
	require.NoError(t, err)

	select {
	case e := <-w.CreateEventChan:
		require.Equal(t, testFileName, e.Name)
		require.Equal(t, nestedDir, e.Path)
		require.Equal(t, testFile, e.FullPath)
		require.Equal(t, File, e.FileType)
		require.Equal(t, Create, e.EventType)
		require.WithinDuration(t, time.Now(), e.ModifiedAt, time.Second)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}
