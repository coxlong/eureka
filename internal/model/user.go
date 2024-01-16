package model

type User struct {
	ID        string `json:"id"`
	OAuthType string `json:"oauth_type"`
	OAuthID   string `json:"oauth_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}
