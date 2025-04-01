package main

import (
	"fmt"
	"log"

	"github.com/GHOSTEDDIE/miservice-go/miservice"
)

func main() {
	client := miservice.NewClient(
		"2901732950",
		"kz185132",
	)
	//devices, err2 := client.ListMinaDevices(1)
	//if err2 != nil {
	//	log.Fatal(err2)
	//}
	//fmt.Println(devices[0])

	//devices, err2 := client.ListMinaDevices(1)
	//if err2 != nil {
	//	log.Fatal(err2)
	//}
	//fmt.Println(devices[0])
	client.MiotAction("936062221", []int{5, 1}, []any{"¿ʞо ∩оʎ ǝɹɐ"})
	//client = miservice.NewClient(
	//	"2901732950",
	//	"kz185132",
	//)
	conversations, err := client.GetConversations("936062221")

	if err != nil {
		return
	}
	fmt.Println(conversations)

	//client.PlayerSetVolume("f1801b9f-0034-4153-bbf6-6b6453668c26", 30)
	//resp, err := client.TextToSpeech("f1801b9f-0034-4153-bbf6-6b6453668c26", "你好，我是小爱")
	if err != nil {
		log.Fatal(err)
	}
}
