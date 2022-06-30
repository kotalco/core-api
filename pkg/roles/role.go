package roles

const (
	Admin  = "admin"
	Reader = "reader"
	writer = "writer"
)

type Roles struct {
}

func New() *Roles {
	return &Roles{}
}

func (r *Roles) Exist(role string) bool {
	if role == Admin || role == Reader || role == writer {
		return true
	}
	return false
}
