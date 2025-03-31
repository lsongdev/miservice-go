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
	//devices, err2 := client.ListMinaDevices(1)
	//if err2 != nil {
	//	log.Fatal(err2)
	//}
	//fmt.Println(devices[0])

	devices, err2 := client.ListMinaDevices(1)
	if err2 != nil {
		log.Fatal(err2)
	}
	for _, device := range devices {
		log.Println(device.Name, device.MiotDID)
	}
	// fmt.Println(devices[0])
	// client.MiotAction("104854757", []int{5, 1}, []any{"hi"})
	// client.Speak("104854757", "hi")
	records, err := client.GetConversations("104854757")
	if err != nil {
		log.Fatal(err)
	}
	for _, record := range records {
		for _, answer := range record.Answers {
			if answer["type"] == "TTS" {
				log.Println(record.Query, "===>", answer["tts"].(map[string]any)["text"])
			} else {
				log.Println(record.Query, answer)
			}
		}
	}
	// client.PlayerSetVolume("f1801b9f-0034-4153-bbf6-6b6453668c26", 30)
	// resp, err := client.TextToSpeech("f1801b9f-0034-4153-bbf6-6b6453668c26", "你好，我是小爱")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println(resp)
}
