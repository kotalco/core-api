package roles

const (
	Admin  = "admin"
	Reader = "reader"
	Writer = "writer"
)

type Roles struct {
}

func New() *Roles {
	return &Roles{}
}

func (r *Roles) Exist(role string) bool {
	if role == Admin || role == Reader || role == Writer {
		return true
	}
	return false
}
