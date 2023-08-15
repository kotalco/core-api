package endpointactivity

type Activity struct {
	ID         string
	EndpointId string `gorm:"uniqueIndex"`
	Counter    int64
}
