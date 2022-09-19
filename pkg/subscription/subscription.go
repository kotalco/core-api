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

func IsValid() bool {
	if SubscriptionDetails == nil || SubscriptionDetails.Status != "active" {
		return false
	}
	return true
}
