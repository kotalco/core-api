package endpointactivity

type Activity struct {
	ID         string `gorm:"uniqueIndex"`
	EndpointId string
	UserId     string
	Timestamp  int64
}
