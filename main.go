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
	client.PlayerSetVolume("f1801b9f-0034-4153-bbf6-6b6453668c26", 30)
	resp, err := client.TextToSpeech("f1801b9f-0034-4153-bbf6-6b6453668c26", "你好，我是小爱")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
}
