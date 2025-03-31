package miservice

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type H = map[string]any

const UA = "MiHome/6.0.103 (com.xiaomi.mihome; build:6.0.103.1; iOS 14.4.0) Alamofire/6.0.103 MICO/iOSApp/appStore/6.0.103"

type SidToken struct {
	Ssecurity    string `json:"ssecurity"`
	ServiceToken string `json:"service_Token"`
}

type Tokens struct {
	UserName  string              `json:"user_name"`
	DeviceId  string              `json:"device_id"`
	UserId    string              `json:"user_id"`
	PassToken string              `json:"pass_Token"`
	Sids      map[string]SidToken `json:"sids"`
}

func NewTokens() *Tokens {
	return &Tokens{
		Sids: make(map[string]SidToken),
	}
}

type Client struct {
	Token    *Tokens
	Client   *http.Client
	Username string
	Password string
}

func NewClient(username string, password string) *Client {
	j, _ := cookiejar.New(nil)
	return &Client{
		Client: &http.Client{
			Jar: j,
		},
		Username: username,
		Password: password,
	}
}

type loginResp struct {
	Qs             string      `json:"qs"`
	Ssecurity      string      `json:"ssecurity"`
	Code           int         `json:"code"`
	PassToken      string      `json:"passToken"`
	Description    string      `json:"description"`
	SecurityStatus int         `json:"securityStatus"`
	Nonce          int64       `json:"nonce"`
	UserID         int         `json:"userId"`
	CUserID        string      `json:"cUserId"`
	Result         string      `json:"result"`
	Psecurity      string      `json:"psecurity"`
	CaptchaURL     interface{} `json:"captchaUrl"`
	Location       string      `json:"location"`
	Pwd            int         `json:"pwd"`
	Child          int         `json:"child"`
	Desc           string      `json:"desc"`

	ServiceParam string `json:"serviceParam"`
	Sign         string `json:"_sign"`
	Sid          string `json:"sid"`
	Callback     string `json:"callback"`
}

// sid: service id, like "xiaomiio", "micoapi", "mina"
func (ma *Client) Login(sid string) error {
	var err error
	if ma.Token == nil {
		ma.Token = NewTokens()
		ma.Token.UserName = ma.Username
		ma.Token.DeviceId = strings.ToUpper(getRandom(16))
	}
	cookies := []*http.Cookie{
		{Name: "sdkVersion", Value: "3.9"},
		{Name: "deviceId", Value: ma.Token.DeviceId},
	}
	if ma.Token.PassToken != "" {
		cookies = append(cookies, &http.Cookie{Name: "userId", Value: ma.Token.UserId})
		cookies = append(cookies, &http.Cookie{Name: "passToken", Value: ma.Token.PassToken})
	}
	var resp *loginResp
	resp, err = ma.serviceLogin(fmt.Sprintf("serviceLogin?sid=%s&_json=true", sid), nil, cookies)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		data := url.Values{
			"_json":    {"true"},
			"qs":       {resp.Qs},
			"sid":      {resp.Sid},
			"_sign":    {resp.Sign},
			"callback": {resp.Callback},
			"user":     {ma.Username},
			"hash":     {strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(ma.Password))))},
		}
		resp, err = ma.serviceLogin("serviceLoginAuth2", data, cookies)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("serviceLoginAuth2 error: %v", resp)
		}
	}
	ma.Token.UserId = fmt.Sprint(resp.UserID)
	ma.Token.PassToken = resp.PassToken

	var serviceToken string
	serviceToken, err = ma.requestServiceToken(resp.Location, resp.Ssecurity, resp.Nonce)
	if err != nil {
		return err
	}
	ma.Token.Sids[sid] = SidToken{
		Ssecurity:    resp.Ssecurity,
		ServiceToken: serviceToken,
	}
	return nil
}

func (ma *Client) serviceLogin(name string, data url.Values, cookies []*http.Cookie) (*loginResp, error) {
	headers := http.Header{
		"User-Agent": []string{UA},
	}
	var reqBody io.Reader
	method := http.MethodGet
	if data != nil {
		reqBody = strings.NewReader(data.Encode())
		method = http.MethodPost
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req, _ := http.NewRequest(method, "https://account.xiaomi.com/pass/"+name, reqBody)
	req.Header = headers
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := ma.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var jsonResponse loginResp
	// "&&&START&&&"
	err = json.Unmarshal(body[11:], &jsonResponse)
	if err != nil {
		return nil, err
	}
	return &jsonResponse, nil
}

func (ma *Client) requestServiceToken(location, ssecurity string, nonce int64) (string, error) {
	nsec := fmt.Sprintf("nonce=%d&%s", nonce, ssecurity)
	sum := sha1.Sum([]byte(nsec))
	clientSign := base64.StdEncoding.EncodeToString(sum[:])
	es := url.QueryEscape(clientSign)
	requestUrl := fmt.Sprintf("%s&clientSign=%s", location, es)
	req, _ := http.NewRequest(http.MethodGet, requestUrl, nil)
	headers := http.Header{
		"User-Agent": []string{UA},
	}
	req.Header = headers
	resp, err := ma.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	cookies := resp.Cookies()
	var serviceToken string
	for _, cookie := range cookies {
		if cookie.Name == "serviceToken" {
			serviceToken = cookie.Value
			break
		}
	}
	if serviceToken == "" {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(string(body))
	}
	return serviceToken, nil
}

type DataCb func(Tokens *Tokens, cookie map[string]string) url.Values

func (ma *Client) NewRequest(sid, u string, data url.Values, cb DataCb, headers http.Header) *http.Request {
	var req *http.Request
	var body io.Reader
	cookies := []*http.Cookie{
		{Name: "userId", Value: ma.Token.UserId},
		{Name: "serviceToken", Value: ma.Token.Sids[sid].ServiceToken},
	}
	method := http.MethodGet
	if data != nil || cb != nil {
		var vals url.Values
		if cb != nil {
			var cookieMap = make(map[string]string)
			vals = cb(ma.Token, cookieMap)
			for k, v := range cookieMap {
				cookies = append(cookies, &http.Cookie{Name: k, Value: v})
			}
		} else if data != nil {
			vals = data
		}
		if vals != nil {
			method = http.MethodPost
			// log.Println("request data", vals.Encode())
			body = strings.NewReader(vals.Encode())
			headers.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	req, _ = http.NewRequest(method, u, body)
	if headers != nil {
		req.Header = headers
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	return req
}

func (ma *Client) hasSid(sid string) bool {
	if ma.Token == nil {
		return false
	}
	_, ok := ma.Token.Sids[sid]
	return ok
}

func (ma *Client) Request(sid, u string, data url.Values, cb DataCb, headers http.Header, output H) error {
	if !ma.hasSid(sid) {
		err := ma.Login(sid)
		if err != nil {
			return err
		}
	}
	req := ma.NewRequest(sid, u, data, cb, headers)
	resp, err := ma.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return err
	}
	if output["code"].(float64) != 0 {
		err = fmt.Errorf("error: %s (code: %2.f)", output["message"], output["code"])
	}
	return err
}
