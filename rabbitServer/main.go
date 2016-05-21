package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	//"strconv"
	//"strings"
	//"time"
	"github.com/streadway/amqp"
)

/*
settings ...
*/
var settings struct {
	DateTimeFormatString, LogFilePath, LogFileName, LogFileSeparator, MQconnectStr string
	StdOut, FileOut, StdOutTrace, StdOutInfo, StdOutWarn, StdOutError              bool
	FileOutMethod                                                                  string
}

/*
Msg ...
*/
type Msg struct {
	MsgType    string
	DtTimeStr  string
	ErrCode    string
	AppName    string
	PkgName    string
	ModuleName string
	FuncName   string
	Line       string
	LogText    string
	LogContext string
}

/*
failOnError ...
*/
func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s\n", msg, err)
		panic(fmt.Sprintf("%s: %s\n", msg, err))
	}
}

/*
init ...
*/
func init() {
	absPath, err := filepath.Abs("./config.json")
	failOnError(err, "ошибка получения абсолютного пути к файлу конфигурации")
	//fmt.Println("try to load config.json from:", absPath)
	configFile, err := os.Open(absPath)
	failOnError(err, "ошибка чтения файла конфигурации")
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&settings)
	failOnError(err, "ошибка парсинга файла конфигурации")
	defer configFile.Close()
}

/*
main ...
*/
func main() {

	conn, err := amqp.Dial(settings.MQconnectStr)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"logs", // name
		false,  // durable
		false,  // delete when usused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			receiveMessage(d.Body)
		}
	}()

	fmt.Printf(" [*] Waiting for messages. To exit press CTRL+C\n\n")
	<-forever

}

func receiveMessage(msgByteArr []byte) {
	// сначала распарсим в структуру
	var message Msg
	err := json.Unmarshal(msgByteArr, &message)
	failOnError(err, "ошибка парсинга сообщения из json в структуру")
	//
	// выведем в консоль если стоит в настройках
	if settings.StdOut == true {
		sendMsgToStdout(message)
	}
	// выведем в текстовый файл если стоит в настройках
	if settings.FileOut == true {
		writeMsgToFile(message, settings.FileOutMethod)
	}
}

/*
sendMsgToStdout ...
*/
func sendMsgToStdout(message Msg) {
	//msgType := message.MsgType
	// ... todo
}

/*
writeMsgToFile ...
*/
func writeMsgToFile(message Msg, method string) {
	//msgType := message.MsgType
	// ... todo
}
