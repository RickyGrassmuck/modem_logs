package mb8611

import (
	"encoding/json"
)

type LoginRequest struct {
	SOAPAction string `json:"-"`
	Login      struct {
		Action        string `json:"Action"`
		Username      string `json:"Username"`
		LoginPassword string `json:"LoginPassword"`
		Captcha       string `json:"Captcha"`
		PrivateLogin  string `json:"PrivateLogin"`
	} `json:"Login"`
	LoginResponse struct {
		Challenge   string `json:"Challenge"`
		Cookie      string `json:"Cookie"`
		PublicKey   string `json:"PublicKey"`
		LoginResult string `json:"LoginResult"`
	} `json:"LoginResponse"`
}

func NewLoginRequest(username, password string) *LoginRequest {
	var req LoginRequest = LoginRequest{}
	req.SOAPAction = "http://purenetworks.com/HNAP1/Login"
	req.Login.Action = "request"
	req.Login.Username = username
	req.Login.LoginPassword = ""
	req.Login.Captcha = ""
	req.Login.PrivateLogin = password
	return &req
}

func (r *LoginRequest) Action() string {
	return r.SOAPAction
}

func (r *LoginRequest) Marshal() []byte {
	ret, _ := json.Marshal(r)
	return ret
}

func (r *LoginRequest) MarshalRequest() []byte {
	ret, _ := json.Marshal(r.Login)
	return ret
}

func (r *LoginRequest) MarshalIndent() []byte {
	ret, _ := json.MarshalIndent(r, "", "  ")
	return ret
}
