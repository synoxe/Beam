package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const DiscoveryMessage = "BEAM_DISCOVER"
const DiscoveryPort = 9999

type Peer struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func StartResponder(sharePort string) error {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: DiscoveryPort,
	}

	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		return err
	}

	go func() {
		defer conn.Close()

		buffer := make([]byte, 2048)

		for {
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Discovery read hatası:", err)
				continue
			}

			message := strings.TrimSpace(string(buffer[:n]))
			if message != DiscoveryMessage {
				continue
			}

			peer := Peer{
				Name: hostName(),
				IP:   localIPv4(),
				Port: sharePort,
			}

			data, err := json.Marshal(peer)
			if err != nil {
				fmt.Println("Discovery marshal hatası:", err)
				continue
			}

			_, err = conn.WriteToUDP(data, remoteAddr)
			if err != nil {
				fmt.Println("Discovery cevap gönderme hatası:", err)
				continue
			}
		}
	}()

	return nil
}

func Discover(timeout time.Duration) ([]Peer, error) {
	broadcastAddr := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: DiscoveryPort,
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = conn.SetWriteBuffer(2048)
	if err != nil {
		return nil, err
	}

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteToUDP([]byte(DiscoveryMessage), &broadcastAddr)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var peers []Peer
	buffer := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return nil, err
		}

		var peer Peer
		err = json.Unmarshal(buffer[:n], &peer)
		if err != nil {
			fmt.Println("Discovery cevap parse hatası:", err)
			continue
		}

		key := peer.IP + ":" + peer.Port
		if seen[key] {
			continue
		}

		seen[key] = true
		peers = append(peers, peer)
	}

	return peers, nil
}

func hostName() string {
	name, err := os.Hostname()
	if err != nil || strings.TrimSpace(name) == "" {
		return "beam-device"
	}
	return name
}

func localIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
			}

			return ip.String()
		}
	}

	return "unknown"
}
