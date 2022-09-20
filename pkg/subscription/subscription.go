package subscription

var (
	SubscriptionDetails *SubscriptionDetailsDto
	CheckDate           int64
	ActivationKey       string
)

type SubscriptionDetailsDto struct {
	Status     string `json:"status"`
	Name       string `json:"name"`
	StartDate  int64  `json:"start_date"`
	EndDate    int64  `json:"end_date"`
	CanceledAt int64  `json:"canceled_at"`
}

var IsValid = func() bool {
	if SubscriptionDetails == nil {
		return false
	} else if !(SubscriptionDetails.Status == "active" || SubscriptionDetails.Status == "past_due") {
		return false
	}
	return true
}
