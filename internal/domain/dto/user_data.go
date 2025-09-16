package dto

type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenData struct {
	Token string `json:"token"`
	Login string `json:"login"`
}
