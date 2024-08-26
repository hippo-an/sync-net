package transfer

import (
	"encoding/binary"
	"fmt"
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/hippo-an/sync-net/pkg/discovery"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Client struct {
	conf *config.Config
	w    *watcher.Watcher
	s    *discovery.Server
	wg   sync.WaitGroup
}

func NewClient(conf *config.Config, w *watcher.Watcher, s *discovery.Server) *Client {
	return &Client{
		conf: conf,
		w:    w,
		s:    s,
	}
}

func (c *Client) HandleEvents() {
	for {
		select {
		case event := <-c.w.CreateEventChan:
			c.handleEvent(event)
		case event := <-c.w.ModifyEventChan:
			c.handleEvent(event)
		case event := <-c.w.DeleteEventChan:
			c.handleEvent(event)
		case err := <-c.w.ErrorChan:
			log.Printf("error from watcher: %s\n", err)
		case <-c.w.StopChan:
			return
		}
	}
}

func (c *Client) handleEvent(event *watcher.Event) {
	for _, serverInfo := range c.s.ServerInfos {

		if serverInfo.Self {
			continue
		}

		c.wg.Add(1)

		go func(s *discovery.ServerInfo) {
			defer c.wg.Done()
			log.Println("handshake with server: ", s.Ip)
			conn, err := net.Dial("tcp", net.JoinHostPort(s.Ip, s.Port))
			if err != nil {
				log.Printf("Failed to connect to server %+v: %s\n", s, err)
				return
			}
			defer conn.Close()

			err = c.handshake(conn, fmt.Sprintf("%d:%s", event.EventType, event.FullPath))
			if err != nil {
				log.Printf("Failed to handshake with server %+v: %s\n", s, err)
				return
			}

			if event.EventType != watcher.Delete {
				err := c.fileTransfer(conn, event.Name)
				if err != nil {
					log.Printf("Failed to send file %+v: %s\n", s, err)
					return
				}
			}
		}(serverInfo)
	}

	c.wg.Wait()
	log.Println("File successfully sent to all servers.")
}

func (c *Client) fileTransfer(conn net.Conn, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Failed to open file %s: %s\n", fileName, err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, c.conf.Transfer.BufferSize)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading file %s: %s\n", fileName, err)
				return err
			}

			return nil
		}

		if n == 0 {
			break
		}

		_, err = conn.Write(buffer[:n])

		if err != nil {
			log.Printf("Failed to send file %s to server: %s\n", fileName, err)
			return err
		}
	}

	log.Printf("Successfully sent file %s to server.\n", fileName)
	return nil
}

func (c *Client) handshake(conn net.Conn, message string) error {
	size := int64(len(message))
	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, uint64(size))

	_, err := conn.Write(sizeBytes)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(message))
	if err != nil {
		return err
	}

	return nil
}
