package real_time

type EventMessage struct {
	ID                string `json:"id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderDeviceID    string `json:"sender_device_id"`
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientDeviceID string `json:"recipient_device_id"`
	CipherText        string `json:"cipher_text"`
	PubKey            string `json:"pub_key"`
	CreatedAt         string `json:"created_at"`
}

type EventGroupMessage struct {
	ID                string `json:"id"`
	GroupID           string `json:"group_id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderDeviceID    string `json:"sender_device_id"`
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientDeviceID string `json:"recipient_device_id"`
	CipherText        string `json:"cipher_text"`
	PubKey            string `json:"pub_key"`
	CreatedAt         string `json:"created_at"`
}

type EventTyping struct {
	UserID  string `json:"user_id"`
	PeerID  string `json:"peer_id"`
	Mode    string `json:"mode"` // "start" | "stop"
	IsGroup bool   `json:"is_group"`
}

type EventStatus struct {
	UserID string `json:"user_id"`
	Status string `json:"status"` // "online" | "offline"
}

type EventAckDelivered struct {
	MessageID  string `json:"message_id"`
	ToUserID   string `json:"to_user_id"`
	ToDeviceID string `json:"to_device_id"`
}

type EventAckRead struct {
	MessageID  string `json:"message_id"`
	ToUserID   string `json:"to_user_id"`
	ToDeviceID string `json:"to_device_id"`
}
