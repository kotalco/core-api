package workspace

import "github.com/kotalco/core-api/core/workspaceuser"

type Workspace struct {
	ID             string
	Name           string
	K8sNamespace   string                        `gorm:"<-:create;uniqueIndex"` //allow read and create only
	UserId         string                        `gorm:"index"`
	WorkspaceUsers []workspaceuser.WorkspaceUser `gorm:"constraint:OnDelete:CASCADE"`
}
