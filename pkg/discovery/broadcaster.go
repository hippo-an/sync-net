package discovery

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

type Broadcaster struct {
	Addr *net.UDPAddr
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Addr: &net.UDPAddr{
			Port: broadcastPort,
			IP:   net.IPv4bcast,
		},
	}
}

func (b *Broadcaster) Broadcast() {
	conn, err := net.DialUDP("udp", nil, b.Addr)
	if err != nil {
		log.Println("Error setting up UDP connection:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Minute)
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
