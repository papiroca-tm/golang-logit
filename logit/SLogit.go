package logit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

var settings struct {
	DateTimeFormatString                                              string
	StackLevelTrace, StackLevelInfo, StackLevelWarn, StackLevelError  int
	StdOut, FileOut, StdOutTrace, StdOutInfo, StdOutWarn, StdOutError bool
}

// Msg ...
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

func init() {
	absPath, _ := filepath.Abs("../github.com/papiroca-tm/golang-logit/logit/config.json")
	fmt.Println("try to load config.json from:", absPath)
	configFile, err := os.Open(absPath)
	failOnError(err, "ошибка чтения файла конфигурации")
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&settings)
	failOnError(err, "ошибка парсинга файла конфигурации")
}

// sendMessage ...
func sendMessage(msgType string, stackLevel int, logContext string, logText string, errCode string) {
	var message Msg
	message.MsgType = msgType
	message.DtTimeStr = timeToStr(time.Now())
	message.ErrCode = errCode
	message.AppName = getAppName()
	message.PkgName = getPkgName(stackLevel)
	message.ModuleName = getModuleName(stackLevel)
	message.FuncName = getFuncName(stackLevel)
	message.Line = getLine(stackLevel)
	message.LogText = logText
	message.LogContext = logContext
	// выведем в консоль если стоит в настройках
	if settings.StdOut == true {
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
	// выведем в текстовый файл если стоит в настройках
	if settings.FileOut == true {
		// todo пишем лог в файл
	}
	// отправляем месседж серверу
	jsonMsg, err := msgToJSON(message)
	failOnError(err, "ошибка msgToJSON(message)")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"logs", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")
	body := jsonMsg
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	failOnError(err, "Failed to publish a message")
}

// stdPrint ...
func stdPrint(fmtStr string, msg Msg) {
	fmt.Printf(fmtStr, msg.MsgType, msg.DtTimeStr, msg.ErrCode, msg.AppName, msg.PkgName, msg.ModuleName, msg.FuncName, msg.Line, msg.LogText)
}

// failOnError ...
func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s\n", msg, err)
		panic(fmt.Sprintf("%s: %s\n", msg, err))
	}
}

// strToTime ...
func strToTime(s string) time.Time {
	t, err := time.Parse(settings.DateTimeFormatString, s)
	failOnError(err, "ошибка парсинга времени из строки в time.Time")
	return t
}

// timeToStr ...
func timeToStr(t time.Time) string {
	return t.Format(settings.DateTimeFormatString)
}

// getAppName ...
func getAppName() string {
	appName := strings.Split(os.Args[0], "/")
	return appName[len(appName)-1]
}

// getPkgName ...
func getPkgName(stackLevel int) string {
	pc, _, _, _ := runtime.Caller(stackLevel)
	functionObject := runtime.FuncForPC(pc)
	arr := strings.Split(functionObject.Name(), ".")
	sPkg := strings.Split(arr[0], "/")
	return sPkg[len(sPkg)-1]
}

// getModuleName ...
func getModuleName(stackLevel int) string {
	_, modulePathName, _, _ := runtime.Caller(stackLevel)
	sModule := strings.Split(modulePathName, "/")
	return sModule[len(sModule)-1]
}

// getFuncName ...
func getFuncName(stackLevel int) string {
	pc, _, _, _ := runtime.Caller(stackLevel)
	functionObject := runtime.FuncForPC(pc)
	arr := strings.Split(functionObject.Name(), ".")
	return arr[len(arr)-1]
}

// getLine ...
func getLine(stackLevel int) string {
	_, _, line, _ := runtime.Caller(stackLevel)
	return strconv.Itoa(line)
}

// msgToJSON ...
func msgToJSON(m Msg) (string, error) {
	jsonMsg, err := json.Marshal(m)
	failOnError(err, "ошибка парсинга сообщения в JSON")
	return string(jsonMsg), nil
}

// TRACE ...
func TRACE(logText, logContext string) {
	sendMessage("TRACE", settings.StackLevelTrace, logContext, logText, "")
}

// INFO ...
func INFO(logText, logContext string) {
	sendMessage("INFO", settings.StackLevelInfo, logContext, logText, "")
}

// WARN ...
func WARN(logText, logContext string) {
	sendMessage("WARN", settings.StackLevelWarn, logContext, logText, "")
}

// ERROR ...
func ERROR(logText, logContext, errCode string) {
	sendMessage("ERROR", settings.StackLevelError, logContext, logText, errCode)
}


// selfRecover ...
// func selfRecover() (err error) {
// 	fName := getFuncName(2)
// 	r := recover()
// 	if r != nil {
// 		fmt.Println(fName, r)
// 	}
// 	err = fmt.Errorf("PANIC %s", r)
// 	return err
// }