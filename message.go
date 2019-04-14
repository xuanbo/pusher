package pusher

type MessageType int
type MediaType int

const (
	Single MessageType = iota
	Group
	SysNotify
	OnlineNotify
	OfflineNotify
)

const (
	Text MediaType = iota
	Image
	File
)

// websocket message
type Message struct {
	MessageType MessageType `json:"messageType"`
	MediaType   MediaType   `json:"mediaType"`
	From        string      `json:"from"`
	To          string      `json:"to"`
	Content     string      `json:"content,omitempty"`
	FileId      string      `json:"fileId,omitempty"`
	Url         string      `json:"url,omitempty"`
	CreateAt    int64       `json:"createAt,omitempty"`
	UpdateAt    int64       `json:"updateAt,omitempty"`
}
