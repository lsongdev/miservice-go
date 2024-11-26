# XiaoMi Cloud Service

```go
package main

import (
  "log"
  "os"

  "github.com/lsongdev/miservice-go/miservice"
)

func main() {
  client := miservice.NewClient(
    os.Getenv("MI_USERNAME"),
    os.Getenv("MI_PASSWORD"),
  )
```

*[[repo]/miservice/miio.go](./miservice/miio.go)*

#### MiioRequest

+ HomeRequest
+ [HomeDeviceList](#homedevicelist)
+ HomeGetProps
+ HomeSetProps

#### HomeDeviceList

```go
devices, err := client.HomeDeviceList(&miservice.DeviceListFilter{
  ShowVirtualModel: true,
  ShowHuamiDevices: 1,
})
if err != nil {
  log.Fatal(err)
}
for _, device := range devices {
  log.Println(device.Name, device.Model, device.Did, device.Token)
}
```

*[[repo]/miservice/miot.go](./miservice/miot.go)*

+ MiotGetProps
+ MiotSetProps
+ [MiotAction](#miotaction)

#### MiotAction

<https://miot-spec.org/miot-spec-v2/instances?status=all>

```go
func (c *Client) Speak(did, text string) (H, error) {
  return c.MiotAction(did, []int{5, 1}, []any{text})
}

func (c *Client) WakeUp(did string) (H, error) {
  return c.MiotAction(did, []int{5, 2}, []any{})
}

func (c *Client) PlayRadio(did string) (H, error) {
  return c.MiotAction(did, []int{5, 3}, []any{})
}

func (c *Client) PlayMusic(did string) (H, error) {
  return c.MiotAction(did, []int{5, 4}, []any{})
}
```

*[[repo]/miservice/mina.go](./miservice/mina.go)*

+ MinaRequest
+ [ListMinaDevices](#listminadevices)
+ [RemoteUbusCall](#remoteubuscall)

#### ListMinaDevices

```go
devices, _ := client.ListMinaDevices(1)
for _, device := range devices {
  log.Println(device.Name, device.DeviceID)
}
```

#### RemoteUbusCall

```go
func (c *Client) TextToSpeech(deviceId, text string) (out H, err error) {
  out = make(map[string]any)
  err = c.RemoteUbusCall(deviceId, "mibrain.text_to_speech", H{"text": text}, out)
  return
}
```

```go
did := "f1801b9f-0034-4153-bbf6-6b6453668c26"
client.PlayerSetVolume(did, 30)
resp, err := client.TextToSpeech(did, "你好，我是小爱")
if err != nil {
  log.Fatal(err)
}
log.Println(resp)
```
