package protocol

const (
	MessageTypeFileInfo = "FILE_INFO"
	MessageTypeAccept   = "ACCEPT"
	MessageTypeReject   = "REJECT"
	MessageTypeDone     = "DONE"
	MessageTypeError    = "ERROR"
)

type FileMetadata struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

type Message struct {
	Type     string        `json:"type"`
	Metadata *FileMetadata `json:"metadata,omitempty"`
	Error    string        `json:"error,omitempty"`
}
