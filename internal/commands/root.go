package commands

import (
	"beam/internal/discovery"
	"beam/internal/transfer"
	"flag"
	"fmt"
	"os"
	"time"
)

func Commands() {
	if len(os.Args) < 2 {
		help()
		return
	}

	command := os.Args[1]

	switch command {
	case "help":
		help()
	case "discover":
		handleDiscover()
	case "send":
		handleSend()
	case "receive":
		handleReceive()

	default:
		fmt.Println("Bilinmeyen komut", command)
		help()
	}
}

func handleDiscover() {
	fmt.Println("Ağda Beam cihazları aranıyor...")

	peers, err := discovery.Discover(2 * time.Second)
	if err != nil {
		fmt.Println("Discovery hatası:", err)
		return
	}

	if len(peers) == 0 {
		fmt.Println("Hiç cihaz bulunamadı.")
		return
	}

	fmt.Println("Bulunan cihazlar:")
	for i, peer := range peers {
		fmt.Printf("%d. %s - %s:%s\n", i+1, peer.Name, peer.IP, peer.Port)
	}
}

func handleSend() {
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	file := sendCmd.String("file", "", "gönderilecek dosya yolu")
	to := sendCmd.String("to", "", "hedef ip:port")

	err := sendCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("send argümanları okunamadı", err)
		return
	}
	if *file == "" || *to == "" {
		fmt.Println("Kullanım: beam send --file dosya.txt --to 192.168.1.25:9000")
		return
	}

	fmt.Println("Send modu çalıştı")
	fmt.Println("Dosya:", *file)
	fmt.Println("Hedef:", *to)

	err = transfer.SendFile(*file, *to)
	if err != nil {
		fmt.Println("Gönderim Hatası", err)
	}

}

func handleReceive() {
	receiveCmd := flag.NewFlagSet("receive", flag.ExitOnError)

	port := receiveCmd.String("port", "9000", "dinlenecek port")
	dir := receiveCmd.String("dir", "downloads", "dosyaların kaydedileceği klasör")

	err := receiveCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("Receive argümanları okunamadı:", err)
		return
	}

	fmt.Println("Receive modu çalıştı")
	fmt.Println("Port:", *port)
	fmt.Println("Kayıt klasörü:", *dir)

	err = discovery.StartResponder(*port)
	if err != nil {
		fmt.Println("Discovery responder başlatılamadı:", err)
		return
	}

	err = transfer.StartReceiver(*port, *dir)
	if err != nil {
		fmt.Println("Receiver başlatılamadı:", err)
	}
}

func help() {
	fmt.Println("--------------------------------------------------")
	fmt.Println("Beam - Aynı ağdaki kullanıcılarla dosya paylaşımı")
	fmt.Println()
	fmt.Println("Kullanım:")
	fmt.Println("  beam help")
	fmt.Println("  beam discover")
	fmt.Println("  beam receive --port 9000")
	fmt.Println("  beam send --file dosya.txt --to 192.168.1.25:9000")
	fmt.Println("--------------------------------------------------")
}
