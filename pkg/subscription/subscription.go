package subscription

var (
	SubscriptionDetails *SubscriptionDetailsDto
	CheckDate           int64
)

type SubscriptionDetailsDto struct {
	Status       string `json:"status"`
	Name         string `json:"name"`
	StartDate    int64  `json:"start_date"`
	EndDate      int64  `json:"end_date"`
	CanceledAt   int64  `json:"canceled_at"`
	TrialStartAt int64  `json:"trial_start_at"`
	TrialEndAt   int64  `json:"trial_end_at"`
	NodesLimit   uint   `json:"nodes_limit"`
}

var IsValid = func() bool {
	if SubscriptionDetails == nil {
		return false
	} else if !(SubscriptionDetails.Status == "active" || SubscriptionDetails.Status == "past_due" || SubscriptionDetails.Status == "trialing") {
		Reset()
		return false
	}
	return true
}

func Reset() {
	SubscriptionDetails = nil
	CheckDate = 0
}
