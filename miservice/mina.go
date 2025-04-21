package miservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (c *Client) MinaRequest(path string, data url.Values, out H) error {
	requestId := "app_ios_" + getRandom(30)
	if data != nil {
		data["requestId"] = []string{requestId}
	} else {
		path += "&requestId=" + requestId
	}
	headers := http.Header{
		"User-Agent": []string{UA},
	}
	return c.Request("micoapi", "https://api2.mina.mi.com"+path, data, nil, headers, out)
}

type DeviceData struct {
	DeviceID     string `json:"deviceID"`
	SerialNumber string `json:"serialNumber"`
	Name         string `json:"name"`
	Alias        string `json:"alias"`
	Current      bool   `json:"current"`
	Presence     string `json:"presence"`
	Address      string `json:"address"`
	MiotDID      string `json:"miotDID"`
	Hardware     string `json:"hardware"`
	RomVersion   string `json:"romVersion"`
	Capabilities struct {
		ChinaMobileIms      int `json:"china_mobile_ims"`
		SchoolTimetable     int `json:"school_timetable"`
		NightMode           int `json:"night_mode"`
		UserNickName        int `json:"user_nick_name"`
		PlayerPauseTimer    int `json:"player_pause_timer"`
		DialogH5            int `json:"dialog_h5"`
		ChildMode2          int `json:"child_mode_2"`
		ReportTimes         int `json:"report_times"`
		AlarmVolume         int `json:"alarm_volume"`
		AiInstruction       int `json:"ai_instruction"`
		ClassifiedAlarm     int `json:"classified_alarm"`
		AiProtocol30        int `json:"ai_protocol_3_0"`
		NightModeDetail     int `json:"night_mode_detail"`
		ChildMode           int `json:"child_mode"`
		BabySchedule        int `json:"baby_schedule"`
		ToneSetting         int `json:"tone_setting"`
		Earthquake          int `json:"earthquake"`
		AlarmRepeatOptionV2 int `json:"alarm_repeat_option_v2"`
		XiaomiVoip          int `json:"xiaomi_voip"`
		NearbyWakeupCloud   int `json:"nearby_wakeup_cloud"`
		FamilyVoice         int `json:"family_voice"`
		BluetoothOptionV2   int `json:"bluetooth_option_v2"`
		Yunduantts          int `json:"yunduantts"`
		MicoCurrent         int `json:"mico_current"`
		VoipUsedTime        int `json:"voip_used_time"`
	} `json:"capabilities"`
	RemoteCtrlType  string `json:"remoteCtrlType"`
	DeviceSNProfile string `json:"deviceSNProfile"`
	DeviceProfile   string `json:"deviceProfile"`
	BrokerEndpoint  string `json:"brokerEndpoint"`
	BrokerIndex     int    `json:"brokerIndex"`
	Mac             string `json:"mac"`
	Ssid            string `json:"ssid"`
}

func (c *Client) ListMinaDevices(master int) (devices []DeviceData, err error) {
	var res H = make(map[string]any)
	err = c.MinaRequest(fmt.Sprintf("/admin/v2/device_list?master=%d", master), nil, res)
	if err != nil {
		return nil, err
	}
	deviceList := res["data"].([]any)
	devices = make([]DeviceData, len(deviceList))
	for i, item := range deviceList {
		ja, _ := json.Marshal(item)
		json.Unmarshal(ja, &devices[i])
	}
	return devices, nil
}

func (c *Client) FindDeviceById(did string) (device DeviceData, err error) {
	devices, err := c.ListMinaDevices(1)
	if err != nil {
		return
	}
	for _, d := range devices {

		if d.MiotDID == did {
			device = d
			return
		}
	}
	return
}

type ConversationRecord struct {
	Query     string           `json:"query"`
	Answers   []map[string]any `json:"answers"`
	Time      int64            `json:"time"`
	RequestID string           `json:"requestId"`
}

func (c *Client) GetConversations(did string) (records []ConversationRecord, err error) {
	c.Login("micoapi")
	device, err := c.FindDeviceById(did)
	if err != nil {
		return
	}
	guid := uuid.New().String()
	url := fmt.Sprintf("https://userprofile.mina.mi.com/device_profile/v2/conversation?limit=%d&requestId=%s&source=dialogu&hardware=%s", 10, guid, device.Hardware)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", UA)
	req.Header.Add("Referer", "ttps://userprofile.mina.mi.com/dialogue-note/index.html")
	req.Header.Add("Cookie", fmt.Sprintf(`userId=%s; serviceToken=%s; deviceId=%s`, c.Token.UserId, c.Token.Sids["micoapi"].ServiceToken, device.DeviceID))
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	var m struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	err = json.Unmarshal(body, &m)
	if m.Code != 0 {
		err = errors.New(m.Message)
		return
	}
	var k struct {
		Records []ConversationRecord `json:"records"`
	}
	err = json.Unmarshal([]byte(m.Data), &k)
	records = k.Records
	return
}

func (c *Client) RemoteUbusCall(deviceId, cmd string, message H, res H) error {
	parts := strings.Split(cmd, ".")
	path := parts[0]
	method := parts[1]
	messageJSON, _ := json.Marshal(message)
	data := url.Values{
		"deviceId": []string{deviceId},
		"message":  []string{string(messageJSON)},
		"method":   []string{method},
		"path":     []string{path},
	}
	return c.MinaRequest("/remote/ubus", data, res)
}

func (c *Client) TextToSpeech(deviceId, text string) (out H, err error) {
	out = make(map[string]any)
	err = c.RemoteUbusCall(deviceId, "mibrain.text_to_speech", H{"text": text}, out)
	return
}

func (c *Client) PlayerSetVolume(deviceId string, volume int) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_set_volume", H{"volume": volume, "media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerGetStatus(deviceId string) (any, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_get_play_status", H{"media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerPlay(deviceId string) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_play_operation", H{"action": "play", "media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerPause(deviceId string) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_play_operation", H{"action": "pause", "media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerResume(deviceId string) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_play_operation", H{"action": "resume", "media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerStop(deviceId string) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_play_operation", H{"action": "stop", "media": "app_ios"}, res)
	return res, err
}

func (c *Client) PlayerPlayUrl(deviceId, url string) (H, error) {
	var res H = make(map[string]any)
	err := c.RemoteUbusCall(deviceId, "mediaplayer.player_play_url", H{"url": url, "type": 2, "media": "app_ios"}, res)
	return res, err
}
