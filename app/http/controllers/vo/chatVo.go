package vo

// ChatItemVo summarizes one conversation in the messages list.
type ChatItemVo struct {
	Id           uint64 `json:"id"` // user_chat_config id
	PeerId       uint64 `json:"peerId"`
	PeerUsername string `json:"peerUsername"`
	PeerAvatar   string `json:"peerAvatar"`
	LastMsg      string `json:"lastMsg"`
	LastMsgTime  string `json:"lastMsgTime"`
	UnreadCount  uint   `json:"unreadCount"`
	ConvId       uint64 `json:"convId"`
}

// MessageVo represents one chat message decorated for the current viewer.
type MessageVo struct {
	Id        uint64 `json:"id"`
	SenderId  uint64 `json:"senderId"`
	Content   string `json:"content"`
	MsgType   int8   `json:"msgType"`
	IsRead    int    `json:"isRead"`
	CreatedAt string `json:"createdAt"`
	IsSelf    bool   `json:"isSelf"`
}
