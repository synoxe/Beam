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

func SendFile(path string, address string) error {
	checksum, err := hashutil.FileSHA256(path)
	if err != nil {
		return fmt.Errorf("checksum hesaplanamadı: %w", err)
	}

	meta, err := storage.BuildMetadata(path, checksum)
	if err != nil {
		return fmt.Errorf("metadata oluşturulamadı: %w", err)
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("hedefe bağlanılamadı: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	msg := protocol.Message{
		Type:     protocol.MessageTypeFileInfo,
		Metadata: meta,
	}

	err = json.NewEncoder(writer).Encode(msg)
	if err != nil {
		return fmt.Errorf("metadata gönderilemedi: %w", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("buffer flush edilemedi: %w", err)
	}

	var response protocol.Message
	err = json.NewDecoder(reader).Decode(&response)
	if err != nil {
		return fmt.Errorf("receiver cevabı okunamadı: %w", err)
	}

	if response.Type != protocol.MessageTypeAccept {
		if response.Error != "" {
			return fmt.Errorf("receiver reddetti: %s", response.Error)
		}
		return fmt.Errorf("receiver dosyayı kabul etmedi")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	tracker := progress.NewTracker("Sending", meta.Size)
	tracker.Start()

	_, err = io.Copy(io.MultiWriter(writer, tracker), file)
	if err != nil {
		return fmt.Errorf("dosya verisi gönderilemedi: %w", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("dosya verisi flush edilemedi: %w", err)
	}

	tracker.Finish()

	err = json.NewDecoder(reader).Decode(&response)
	if err != nil {
		return fmt.Errorf("transfer sonucu okunamadı: %w", err)
	}

	if response.Type != protocol.MessageTypeDone {
		if response.Error != "" {
			return fmt.Errorf("transfer başarısız: %s", response.Error)
		}
		return fmt.Errorf("transfer tamamlanmadı")
	}

	fmt.Println("Dosya başarıyla gönderildi:", meta.FileName)
	return nil
}
