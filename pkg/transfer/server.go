package transfer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	conf *config.Config
}

func NewServer(conf *config.Config) *Server {
	return &Server{
		conf: conf,
	}
}

func (s *Server) ListenAndConnect(port int) {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Error starting server:", err)
	}

	for {
		conn, err := conn.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handleConnection(conn)

	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	sizeBytes := make([]byte, 8)
	_, err := conn.Read(sizeBytes)
	if err != nil {
		log.Println("Error reading size:", err)
		return
	}

	size := binary.BigEndian.Uint64(sizeBytes)

	buffer := make([]byte, size)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading message:", err)
		return
	}

	eventInfo := string(buffer[:n])
	eventType, filePath, err := parseEventInfo(eventInfo)
	if err != nil {
		log.Println("Error parsing event info:", err)
		return
	}

	switch eventType {
	case watcher.Create:
		err := s.handleCreateEvent(conn, filePath)
		if err != nil {
			log.Println("Error handling create event:", err)
		}
	case watcher.Modify:
		err := s.handleModifyEvent(conn, filePath)
		if err != nil {
			log.Println("Error handling modify event:", err)
		}
	case watcher.Delete:
		err := s.handleDeleteEvent(filePath)
		if err != nil {
			log.Println("Error handling delete event:", err)
		}
	default:
		log.Printf("Unknown event type: %d", eventType)
	}

}

func parseEventInfo(info string) (watcher.EventType, string, error) {
	parts := strings.SplitN(info, ":", 2)
	if len(parts) != 2 {
		return 0, "", errors.New("invalid event info")
	}

	eventType, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", errors.New("invalid event type")
	}

	return watcher.EventType(eventType), parts[1], nil
}

func (s *Server) handleCreateEvent(conn net.Conn, filePath string) error {
	log.Println("Received file create event for:", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	buf := make([]byte, s.conf.Transfer.BufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from connection:", err)
				return err
			}
			return nil
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			log.Println("Error writing to file:", err)
			return err
		}
	}
}

func (s *Server) handleModifyEvent(conn net.Conn, filePath string) error {
	log.Println("Received file modify event for:", filePath)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	buf := make([]byte, s.conf.Transfer.BufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from connection:", err)
				return err
			}
			break
		}

		if n == 0 {
			log.Println("No more data to read. Ending loop.")
			break
		}

		if n > 0 {
			_, err = file.Write(buf[:n])
			if err != nil {
				log.Println("Error writing to file:", err)
				return err
			}
		}
	}

	return nil
}

func (s *Server) handleDeleteEvent(filePath string) error {
	log.Println("Received file delete event for:", filePath)

	err := os.Remove(filePath)
	if err != nil {
		log.Println("Error deleting file:", err)
		return err
	}

	return nil
}
