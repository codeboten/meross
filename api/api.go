package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type apiResponseLogin struct {
	APIStatus    int               `json:"apiStatus"`
	SystemStatus int               `json:"sysStatus"`
	Data         map[string]string `json:"data"`
	Info         string            `json:"info"`
}

type apiResponseDevices struct {
	APIStatus    int      `json:"apiStatus"`
	SystemStatus int      `json:"sysStatus"`
	Devices      []Device `json:"data"`
	Info         string   `json:"info"`
}

type payload struct {
	Parameters string `json:"params"`
	Signature  string `json:"sign"`
	Timestamp  int64  `json:"timestamp"`
	Nonce      string `json:"nonce"`
}

func (c *Client) baseURL() string {
	if len(c.BaseURL) == 0 {
		return merossURL
	}
	return c.BaseURL
}

func (c *Client) devicesCollectionEndpoint() string {
	return fmt.Sprintf("%s/%s", c.baseURL(), "v1/Device/devList")
}

func (c *Client) loginEndpoint() string {
	return fmt.Sprintf("%s/%s", c.baseURL(), "v1/Auth/Login")
}

// Client is used to connect to the Meross Client
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Logger     *log.Logger
	Key        string
	token      string
	email      string
	password   string
	userID     string
}

// NewClient returns an API with sensible defaults
func NewClient(email, password string) Client {
	return Client{
		HTTPClient: &http.Client{},
		Logger:     log.New(&bytes.Buffer{}, "logger: ", log.Lshortfile),
		email:      email,
		password:   password,
	}
}

func (c *Client) getLogin() map[string]string {
	return map[string]string{
		"email":    c.email,
		"password": c.password,
	}
}

func (c *Client) getSignedPayload(params map[string]string) payload {
	ts := time.Now().UnixNano() / int64(time.Millisecond)
	nonce := getNonce(16)
	encodedParams := encodeParameters(params)
	return payload{
		Parameters: encodedParams,
		Signature:  signRequest(ts, nonce, encodedParams),
		Timestamp:  ts,
		Nonce:      nonce,
	}
}

func (c *Client) getError(info string) error {
	switch info {
	case "No login":
		return errNoLogin
	case "Sign check failed":
		return errSignCheckFailed
	case "Lack user":
		return errMissingUser
	case "Lack password":
		return errMissingPassword
	default:
		return nil
	}
}

func (c *Client) addHeaders(req *http.Request) {
	req.Header["Authorization"] = []string{"Basic"}
	if len(c.token) > 0 {
		req.Header["Authorization"] = []string{fmt.Sprintf("Basic %s", c.token)}
	}
	// : "Basic" if self._token is None else "Basic %s" % self._token,
	req.Header["Vender"] = []string{"Meross"}
	req.Header["AppVersion"] = []string{"1.3.0"}
	req.Header["AppLanguage"] = []string{"EN"}
	req.Header["User-Agent"] = []string{"okhttp/3.6.0"}
	req.Header["Content-Type"] = []string{"application/json"}
}

// Login sends a login request to the Meross API and retrieves a token used
// for further authentication.
func (c *Client) Login() error {
	payload := c.getSignedPayload(c.getLogin())
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.loginEndpoint(), bytes.NewBuffer(data))
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned error: %d", resp.StatusCode)
	}
	var response apiResponseLogin
	err = json.NewDecoder(resp.Body).Decode(&response)
	c.Logger.Printf("%v\n", response)
	if err != nil {
		return err
	}
	if err := c.getError(response.Info); err != nil {
		return err
	}
	c.token = response.Data["token"]
	c.Key = response.Data["key"]
	c.userID = response.Data["userid"]

	return nil
}

// GetSupportedDevices connects to the Meross API and returns a slice of Device objects. This
// method can only be called after a successful call to Login.
func (c *Client) GetSupportedDevices() ([]Device, error) {
	payload := c.getSignedPayload(c.getLogin())
	data, err := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.devicesCollectionEndpoint(), bytes.NewBuffer(data))
	c.addHeaders(req)
	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return []Device{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []Device{}, fmt.Errorf("API returned error: %d", resp.StatusCode)
	}
	var response apiResponseDevices
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	if err := c.getError(response.Info); err != nil {
		return response.Devices, errNoLogin
	}

	c.Logger.Printf("%v\n", response)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	c.Logger.Printf("%s\n", body)
	for idx := range response.Devices {
		response.Devices[idx].userID = c.userID
	}
	return response.Devices, nil
}
