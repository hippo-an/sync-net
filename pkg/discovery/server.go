package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hippo-an/sync-net/pkg/config"
	"log"
	"net"
	"os"
	"time"
)

type Server struct {
	ServerInfos map[string]*ServerInfo
	conf        *config.Config
}

func NewServer(conf *config.Config) *Server {
	now := time.Now()

	ip, err := getIp()
	if err != nil {
		log.Fatal("Failed to get ip address:", err)
	}

	s := ServerInfo{
		Id:        uuid.New(),
		Ip:        ip,
		Port:      fmt.Sprint(conf.Discovery.TcpPort),
		CreatedAt: now,
		UpdatedAt: now,
		Self:      true,
	}

	return &Server{
		ServerInfos: map[string]*ServerInfo{
			ip: &s,
		},
		conf: conf,
	}
}

type Message struct {
	Hash string `json:"hash"`
}

type ServerInfo struct {
	Id        uuid.UUID `json:"id"`
	Ip        string    `json:"ip"`
	Port      string    `json:"port"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Self      bool      `json:"self"`
}

func (s *Server) add(ip string) {
	o, ok := s.ServerInfos[ip]
	now := time.Now()

	if !ok {
		nsi := ServerInfo{
			Id:        uuid.New(),
			Ip:        ip,
			CreatedAt: now,
			UpdatedAt: now,
			Self:      false,
		}

		s.ServerInfos[ip] = &nsi
		log.Println("Added server:", ip)
	} else {
		o.UpdatedAt = now
		log.Println("Updated server:", ip)
	}
}

func (s *Server) Listen() {
	addr := net.UDPAddr{
		Port: s.conf.Discovery.BroadcastPort,
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Println("Error setting up UDP server:", err)
		os.Exit(1)
	}

	defer conn.Close()
	log.Println("Listening for broadcast messages...")

	buffer := make([]byte, s.conf.Discovery.BufferSize)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading UDP message:", err)
			continue
		}

		ip, err := validateAddr(clientAddr)
		if err != nil {
			log.Println("Error validating address:", err)
			continue
		}

		var receivedMessage Message
		msg := buffer[:n]

		err = json.Unmarshal(msg, &receivedMessage)
		if err != nil {
			log.Println("Invalid message format, not a valid JSON:", err)
			continue
		}

		if err := validateHash(receivedMessage.Hash); err != nil {
			log.Printf(
				"Invalid hash: %s\n",
				receivedMessage.Hash,
			)
			continue
		}
		s.add(ip)
	}
}

func validateAddr(addr *net.UDPAddr) (string, error) {
	if !addr.IP.IsGlobalUnicast() {
		return "", errors.New("invalid address: is not global uni cast")
	}

	to4 := addr.IP.To4()
	if addr.IP.IsLoopback() || to4 == nil {
		return "", errors.New("invalid address: is not loop back or not an IPv4 address")
	}

	return to4.String(), nil
}

func getIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			to4 := ipNet.IP.To4()
			if to4 != nil {
				return to4.String(), nil
			}
		}
	}

	return "", errors.New("no ip found")
}
