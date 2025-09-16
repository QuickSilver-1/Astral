package authcontroller

type registerRequest struct {
	Login string `json:"login"`
	Pass  string `json:"pswd"`
	Token string `json:"token"`
}

type loginRequest struct {
	Login string `json:"login"`
	Pass  string `json:"pswd"`
}

type registerResponse struct {
	Response struct {
		Login string `json:"login"`
	} `json:"response"`
}

type loginResponse struct {
	Response struct {
		Token string `json:"token"`
	} `json:"response"`
}

type deleteResponse struct {
	Response struct {
		Token bool `json:"token_id"`
	} `json:"response"`
}