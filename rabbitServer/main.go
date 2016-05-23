package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	//"strconv"
	//"strings"
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

/*
receiveMessage ...
*/
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
	msgType := message.MsgType
	formatStr := "%s %s :%s:%s:%s:%s:%s:%s: %s\n"
	switch msgType {
	case "TRACE":
		if settings.StdOutTrace == true {
			stdPrint(formatStr, message)
		}
	case "INFO":
		formatStr = "%s  %s :%s:%s:%s:%s:%s:%s: %s\n"
		if settings.StdOutInfo == true {
			stdPrint(formatStr, message)
		}
	case "WARN":
		formatStr = "%s  %s :%s:%s:%s:%s:%s:%s: %s\n"
		if settings.StdOutWarn == true {
			stdPrint(formatStr, message)
		}
	case "ERROR":
		if settings.StdOutError == true {
			stdPrint(formatStr, message)
		}
	}
}

/*
stdPrint ...
*/
func stdPrint(fmtStr string, msg Msg) {
	fmt.Printf(fmtStr, msg.MsgType, msg.DtTimeStr, msg.ErrCode, msg.AppName, msg.PkgName, msg.ModuleName, msg.FuncName, msg.Line, msg.LogText)
}

/*
writeMsgToFile ...
*/
func writeMsgToFile(message Msg, method string) {
	t := time.Now()
	filePath := settings.LogFilePath + "/" + settings.FileOutMethod + "/" + t.Format("02-01-2006") + "/"
	fileName := settings.LogFileName
	s := settings.LogFileSeparator
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filePath, 0666)
		failOnError(err, "ошибка создания директории")
	}
	if _, err := os.Stat(filePath + fileName); os.IsNotExist(err) {
		_, err := os.Create(filePath + fileName)
		failOnError(err, "ошибка создания файла")
	}

	if method == "text" {
		var formatStr string
		msgType := message.MsgType
		switch msgType {
		case "INFO", "WARN":
			formatStr = "%s " + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s\n"
		default:
			formatStr = "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s" + s + "%s\n"
		}
		str := fmt.Sprintf(
			formatStr,
			message.MsgType,
			message.DtTimeStr,
			message.ErrCode,
			message.AppName,
			message.PkgName,
			message.ModuleName,
			message.FuncName,
			message.Line,
			message.LogText,
			message.LogContext,
		)
		writeStrToFile(filePath+fileName, str)
	}
	if method == "json" {
		str, err := msgToJSON(message)
		failOnError(err, "ошибка msgToJSON(message)")
		writeStrToFile(filePath+fileName, str+"\n")
	}
}

/*
writeStrToFile ...
*/
func writeStrToFile(file string, str string) {
	f, err := os.OpenFile(file, os.O_APPEND, 0666)
	failOnError(err, "ошибка открытия файла для записи")
	n, err := f.WriteString(str)
	failOnError(err, "ошибка записи строки в файл"+string(n))
	defer f.Close()
}

/*
msgToJSON ...
*/
func msgToJSON(m Msg) (string, error) {
	jsonMsg, err := json.Marshal(m)
	failOnError(err, "ошибка парсинга сообщения в JSON")
	return string(jsonMsg), nil
}
