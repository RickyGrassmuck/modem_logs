package mb8611

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

type ModemConfig struct {
	Endpoint string
	Client   *http.Client
}

type APIRequest interface {
	Action() string
	MarshalRequest() []byte
}

func NewClient(address string) (*ModemConfig, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	config := &ModemConfig{
		Endpoint: address,
		Client:   &client,
	}
	return config, nil
}

func (c *ModemConfig) Login(username, password string) (*LoginRequest, error) {
	loginRequest := NewLoginRequest(username, password)
	_, body, err := c.Post(loginRequest)
	_ = json.Unmarshal(body, loginRequest)
	return loginRequest, err
}

func (c *ModemConfig) GetLogs() (*Logs, error) {
	logs := NewLogs()
	_, body, err := c.Post(logs)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(body, &logs.Response)
	return logs, err
}

// Make an HTTP Post request to the endpoint and return the response.
func (c *ModemConfig) Post(r APIRequest) (*http.Response, []byte, error) {
	req, _ := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(r.MarshalRequest()))
	req.Header.Set("SOAPAction", r.Action())
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, respBody, nil
}
