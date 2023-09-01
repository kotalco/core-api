package endpointactivity

type Activity struct {
	ID         string `gorm:"uniqueIndex"`
	EndpointId string
	Timestamp  int64
}
