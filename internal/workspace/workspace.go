package workspace

type Workspace struct {
	ID           string
	Name         string
	K8sNamespace string
	UserId       string `gorm:"index"`
}
