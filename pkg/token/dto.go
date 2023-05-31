package token

type Token struct {
	AccessToken string
	TokenUuid   string //can be used later to store token id in the database (redis for eg) to log all user token out and maintain the state of token
	Expires     int64
	Authorized  bool
}

type AccessDetails struct {
	TokenUuid  string
	UserId     string
	Authorized bool
}

// UserDetails Dto used to get needed user details
type UserDetails struct {
	ID            string
	PlatformAdmin bool
}
