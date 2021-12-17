package tables

import (
	"time"

	"gorm.io/gorm"
)

type AuthNode struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Nodename  string         `gorm:"column:nodename; index" json:"nodename"`
	NodeID    string         `gorm:"column:node_id; uniqueIndex; size:36" json:"node_id"`
	UUID      string         `gorm:"column:uuid; size:36" json:"uuid"`
}
