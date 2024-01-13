package time

const JavascriptISOString = "2006-01-02T15:04:05.999Z07:00"

// Time hold created and updated at information
type Time struct {
	CreatedAt string `json:"createdAt"`
}
