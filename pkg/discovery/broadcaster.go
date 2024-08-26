package discovery

import (
	"encoding/json"
	"github.com/hippo-an/sync-net/pkg/config"
	"log"
	"net"
	"time"
)

type Broadcaster struct {
	addr *net.UDPAddr
	conf *config.Config
}

func NewBroadcaster(conf *config.Config) *Broadcaster {
	return &Broadcaster{
		addr: &net.UDPAddr{
			Port: conf.Discovery.BroadcastPort,
			IP:   net.IPv4bcast,
		},
		conf: conf,
	}
}

func (b *Broadcaster) Broadcast() {
	conn, err := net.DialUDP("udp", nil, b.addr)
	if err != nil {
		log.Println("Error setting up UDP connection:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(b.conf.Discovery.BroadcastInterval)
	defer ticker.Stop()

	err = notify(conn)
	if err != nil {
		log.Fatal("Error notifying initialize broadcasting:", err)
	}
	for {
		select {
		case <-ticker.C:
			err = notify(conn)
			if err != nil {
				log.Println("Error notifying broadcasting:", err)
				continue
			}

			log.Println("Broadcasting successfully")
		}
	}
}

func notify(conn *net.UDPConn) error {
	message := Message{
		Hash: generateHash(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Fatal("message marshaling error: ", err)
		return err
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		log.Println("Error sending discovery message:", err)
		return err
	}

	return nil
}
