package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BootstrapRequest struct {
	Username string  `json:"username"`
	Email    *string `json:"email"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone"`
	Password string  `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        User   `json:"user"`
}
