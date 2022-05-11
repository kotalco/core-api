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

// AuthorizedUser Dto used to define user type that we get from locals in  *.handlers instead of using user struct from users pkg
type AuthorizedUser struct {
	ID               string
	Email            string
	IsEmailVerified  bool
	Password         string
	TwoFactorCipher  string
	TwoFactorEnabled bool
}
