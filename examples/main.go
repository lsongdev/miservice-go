package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lsongdev/miservice-go/miservice"
)

func main() {
	client := miservice.NewClient(
		os.Getenv("MI_USERNAME"),
		os.Getenv("MI_PASSWORD"),
	)
	devices, err2 := client.ListMinaDevices(1)
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Println("Devices:")
	for _, device := range devices {
		fmt.Println(" - ", device.Name, device.MiotDID)
	}

	// device := devices[0]
	// client.MiotAction("104854757", []int{5, 1}, []any{"hi"})
	// client.Speak("104854757", "hi")
	records, err := client.GetConversationRecords("104854757", 100)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Message History")
	for i, record := range records {
		for _, answer := range record.Answers {
			if answer["type"] == "TTS" {
				fmt.Printf("---------------- %d --------------------\n", i)
				fmt.Println(" - Q:", record.Query)
				fmt.Println(" - A:", answer["tts"].(map[string]any)["text"])

			} else {
				fmt.Println(record.Query, answer)
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
