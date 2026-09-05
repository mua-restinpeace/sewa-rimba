package model

import "time"

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type EquipmentItem struct {
	ID             int       `json:"id"`
	CategoryID     *int      `json:"category_id,omitempty"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	DailyRate      float64   `json:"daily_rate"`
	TotalQuantity  int       `json:"total_quantity"`
	Condition_Notes string    `json:"condition_notes,omitempty"`
	PhotoURL       string    `json:"photo_url,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdateAt       time.Time `json:"update_at"`
}

// decoreate EquipmentItem with availabilty computed for specific date of range
type EquipmentAvailability struct {
	EquipmentItem
	AvailableQuantity int `json:"available_quantity"`
}
