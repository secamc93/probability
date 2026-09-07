package domain

type GoogleAuthURLRequest struct {
	RedirectTo string
}

type GoogleAuthURLResponse struct {
	AuthURL string
	State   string
}

type GoogleCallbackRequest struct {
	Code  string
	State string
}

type GoogleLoginResult struct {
	Session     *LoginResponse
	Profile     *GoogleProfile
	NeedsSignup bool
}

type GoogleProfile struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}
