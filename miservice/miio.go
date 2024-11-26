package miservice

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const MIIO_SID = "xiaomiio"

func (s *Client) MiioRequest(path string, data any) (any, error) {
	prepareData := func(token *Tokens, cookies map[string]string) url.Values {
		cookies["PassportDeviceId"] = token.DeviceId
		return signData(path, data, token.Sids[MIIO_SID].Ssecurity)
	}
	headers := http.Header{
		"User-Agent":                 []string{UA},
		"x-xiaomi-protocal-flag-cli": []string{"PROTOCAL-HTTP2"},
	}
	var resp H = make(map[string]any)
	err := s.Request(MIIO_SID, "https://api.io.mi.com/app"+path, nil, prepareData, headers, resp)
	if err != nil {
		return nil, err
	}
	return resp["result"], nil
}

type DeviceInfo struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Did   string `json:"did"`
	Token string `json:"token"`
}

type DeviceListFilter struct {
	ShowVirtualModel bool `json:"getVirtualModel"`
	ShowHuamiDevices int  `json:"getHuamiDevices"`
}

func (s *Client) HomeDeviceList(filter *DeviceListFilter) (devices []DeviceInfo, err error) {
	result, err := s.MiioRequest("/home/device_list", filter)
	if err != nil {
		return nil, err
	}
	deviceList := result.(H)["list"].([]any)
	devices = make([]DeviceInfo, len(deviceList))
	for i, item := range deviceList {
		ja, _ := json.Marshal(item)
		json.Unmarshal(ja, &devices[i])
	}
	return
}

func (s *Client) HomeRequest(did, method string, params any) (H, error) {
	resp, err := s.MiioRequest("/home/rpc/"+did, H{
		"accessKey": "IOS00026747c5acafc2",
		"id":        1,
		"method":    method,
		"params":    params,
	})
	if err != nil {
		return nil, err
	}
	return resp.(H), nil
}

func (s *Client) HomeGetProps(did string, props []string) (H, error) {
	return s.HomeRequest(did, "get_prop", props)
}

func (s *Client) HomeSetProps(did string, props H) (map[string]int, error) {
	results := make(map[string]int, len(props))
	for prop, value := range props {
		result, err := s.HomeSetProp(did, prop, value)
		if err != nil {
			return nil, err
		}
		results[prop] = result
	}
	return results, nil
}

func (s *Client) HomeGetProp(did, prop string) (any, error) {
	results, err := s.HomeGetProps(did, []string{prop})
	if err != nil {
		return nil, err
	}
	return results[prop], nil
}

func (s *Client) HomeSetProp(did, prop string, value any) (int, error) {
	result, err := s.HomeRequest(did, "set_"+prop, value)
	if err != nil {
		return 0, err
	}
	if result["result"] == "ok" {
		return 0, nil
	}
	return -1, nil
}
