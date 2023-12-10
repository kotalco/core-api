package endpointactivity

import "time"

type Activity struct {
	ID         string `gorm:"uniqueIndex"`
	EndpointId string `gorm:"index"`
	UserId     string
	Timestamp  time.Time
}
