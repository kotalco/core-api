package sendgrid

type MailRequestDto struct {
	Email string
	Token string
}

type WorkspaceInvitationMailRequestDto struct {
	Email         string
	WorkspaceName string
	WorkspaceId   string
}
