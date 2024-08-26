package discovery

import (
	"encoding/json"
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

var (
	conf *config.Config
)

func TestMain(m *testing.M) {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	conf = c
	s := NewServer(c)

	go s.Listen()

	time.Sleep(1 * time.Second)
	os.Exit(m.Run())
}

func TestUDPBroadcastHandling(t *testing.T) {

	message := Message{
		Hash: generateHash(),
	}

	jsonData1, err := json.Marshal(message)
	require.NoError(t, err)

	message = Message{
		Hash: "invalid hash please",
	}

	jsonData2, err := json.Marshal(message)
	require.NoError(t, err)

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		Port: conf.Discovery.BroadcastPort,
		IP:   net.IPv4bcast,
	})
	require.NoError(t, err, "failed to create UDP client connection")
	defer conn.Close()

	_, err = conn.Write(jsonData1)
	_, err = conn.Write(jsonData2)
	require.NoError(t, err, "failed to send UDP message")
	time.Sleep(5 * time.Second)
}
