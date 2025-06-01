package response

import "time"

type GetNotificationResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
} 