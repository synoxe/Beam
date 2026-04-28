package transfer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"beam/internal/hashutil"
	"beam/internal/progress"
	"beam/internal/protocol"
	"beam/internal/storage"
)

func StartReceiver(port string, dir string) error {
	address := ":" + port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Println("Receiver başladı. Dinleniyor", address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Bağlantı kabul hatası", err)
		}
		go handleConnection(conn, dir)
	}
}

func handleConnection(conn net.Conn, dir string) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	var msg protocol.Message

	err := json.NewDecoder(reader).Decode(&msg)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "metadata okunamadı",
		})
		fmt.Println("Metadata decode hatası", err)
		return
	}

	if msg.Type != protocol.MessageTypeFileInfo || msg.Metadata == nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeReject,
			Error: "geçersiz ilk mesaj",
		})
		fmt.Println("Geçersiz ilk mesaj alındı")
		return
	}

	meta := msg.Metadata

	if meta.Size < 0 {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeReject,
			Error: "geçersiz dosya boyutu",
		})
		fmt.Println("Geçersiz dosya boyutu")
		return
	}

	cleanName := storage.SanitizeFileName(meta.FileName)

	err = storage.EnsureDir(dir)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "hedef klasör hazırlanamadı",
		})
		fmt.Println("Klasör oluşturma hatası:", err)
		return
	}

	targetPath := storage.UniqueFilePath(dir, cleanName)
	tempPath := targetPath + ".part"

	err = sendMessage(conn, protocol.Message{
		Type: protocol.MessageTypeAccept,
	})
	if err != nil {
		fmt.Println("ACCEPT mesajı gönderilemedi:", err)
		return
	}

	file, err := os.Create(tempPath)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "dosya oluşturulamadı",
		})
		fmt.Println("Temp dosya oluşturma hatası:", err)
		return
	}

	shouldRemoveTemp := true
	defer func() {
		_ = file.Close()
		if shouldRemoveTemp {
			_ = os.Remove(tempPath)
		}
	}()

	tracker := progress.NewTracker("Receiving", meta.Size)
	tracker.Start()

	written, err := io.CopyN(io.MultiWriter(file, tracker), reader, meta.Size)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "dosya verisi alınamadı",
		})
		fmt.Println("Dosya yazma hatası:", err)
		return
	}

	tracker.Finish()

	if written != meta.Size {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "eksik veri alındı",
		})
		fmt.Println("Eksik veri alındı")
		return
	}

	err = file.Close()
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "dosya kapatılamadı",
		})
		fmt.Println("Dosya kapatma hatası:", err)
		return
	}

	currentHash, err := hashutil.FileSHA256(tempPath)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "hash hesaplanamadı",
		})
		fmt.Println("Hash hesaplama hatası:", err)
		return
	}

	if currentHash != meta.Checksum {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "checksum eşleşmedi",
		})
		fmt.Println("Checksum eşleşmedi")
		return
	}

	err = os.Rename(tempPath, targetPath)
	if err != nil {
		_ = sendMessage(conn, protocol.Message{
			Type:  protocol.MessageTypeError,
			Error: "dosya taşınamadı",
		})
		fmt.Println("Dosya rename hatası:", err)
		return
	}

	shouldRemoveTemp = false

	err = sendMessage(conn, protocol.Message{
		Type: protocol.MessageTypeDone,
	})
	if err != nil {
		fmt.Println("DONE mesajı gönderilemedi:", err)
		return
	}

	fmt.Println("Dosya başarıyla alındı:", targetPath)
}

func sendMessage(conn net.Conn, msg protocol.Message) error {
	return json.NewEncoder(conn).Encode(msg)
}
