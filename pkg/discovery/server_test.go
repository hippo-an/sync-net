package discovery

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	s := NewServer()

	go s.Listen()

	time.Sleep(1 * time.Second)
	os.Exit(m.Run())
}

var (
	serverAddr = net.UDPAddr{
		Port: broadcastPort,
		IP:   net.IPv4bcast,
	}
)

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

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	require.NoError(t, err, "failed to create UDP client connection")
	defer conn.Close()

	_, err = conn.Write(jsonData1)
	_, err = conn.Write(jsonData2)
	require.NoError(t, err, "failed to send UDP message")
	time.Sleep(5 * time.Second)
}
