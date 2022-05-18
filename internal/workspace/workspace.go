package workspace

type Workspace struct {
	ID           string
	Name         string
	K8sNamespace string `gorm:"<-:create"` //allow read and create only
	UserId       string `gorm:"index"`
}
