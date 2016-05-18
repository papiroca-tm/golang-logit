package logit

import (
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
	"time"
	"runtime"
	"os"
	"github.com/streadway/amqp"
)

var settings struct {
	DateTimeFormatString  string
	StackLevelTrace, StackLevelInfo, StackLevelWarn, StackLevelError int
	StdOut, FileOut, StdOutTrace, StdOutInfo, StdOutWarn, StdOutError bool
}

// Msg ...
type Msg struct {
	MsgType string
	DtTimeStr string
	ErrCode string
	AppName string
	PkgName string
	ModuleName string
	FuncName string
	Line string
	LogText string
	LogContext string
}

func init() {
	defer selfRecover()
	dir, err := os.Getwd()
	checkErr(err)
	dirToConfig := dir + "/app/services/logit/config.json" // todo path ?
	configFile, err := os.Open(dirToConfig)
	checkErr(err)
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&settings)
	checkErr(err)
}

// sendMessage ...
func sendMessage(msgType string, stackLevel int, logContext string, logText string, errCode string) {
	defer selfRecover()
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
				if settings.StdOutTrace == true {stdPrint(formatStr, message)}
			case "INFO":
				formatStr = "%s  %s :%s:%s:%s:%s:%s:%s: %s\n"
				if settings.StdOutInfo == true {stdPrint(formatStr, message)}
			case "WARN":
				formatStr = "%s  %s :%s:%s:%s:%s:%s:%s: %s\n"
				if settings.StdOutWarn == true {stdPrint(formatStr, message)}
			case "ERROR":
				if settings.StdOutError == true {stdPrint(formatStr, message)}
		}
	}
	// выведем в текстовый файл если стоит в настройках
	if settings.FileOut == true {
		// todo пишем лог в файл
	}
	// отправляем месседж серверу
	jsonMsg, _ := msgToJSON(message) // todo _ => err	
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ") // todo failOnError => self func
	defer conn.Close()	
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()	
	q, err := ch.QueueDeclare(
		"logs", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")	
	body := jsonMsg
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	failOnError(err, "Failed to publish a message")
}

// stdPrint ...
func stdPrint(fmtStr string, msg Msg) {
	defer selfRecover()
	fmt.Printf(fmtStr, msg.MsgType, msg.DtTimeStr, msg.ErrCode, msg.AppName, msg.PkgName, msg.ModuleName, msg.FuncName, msg.Line, msg.LogText)
}

// failOnError ... todo selfRecover or something like this or completly not
func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s\n", msg, err)
		panic(fmt.Sprintf("%s: %s\n", msg, err))
	}
}

// checkErr ... todo selfRecover or something like this or completly not
func checkErr(err error) {
	defer selfRecover()
	if err != nil {
		fmt.Println(err) // todo format output
	}
}

// selfRecover ...
func selfRecover() {
	fName := getFuncName(2)
	if r := recover(); r != nil {
		fmt.Println(fName, r) // todo format output
	}
}

// strToTime ...
func strToTime(s string) time.Time {
	defer selfRecover()
	t, err := time.Parse(settings.DateTimeFormatString, s)
	checkErr(err)
	return t
}

// timeToStr ...
func timeToStr(t time.Time) string {
	defer selfRecover()
	return t.Format(settings.DateTimeFormatString)
}

// getAppName ...
func getAppName() string {
	// todo try cach
	appName := strings.Split(os.Args[0], "/")
	return appName[len(appName)-1]
}

// getPkgName ...
func getPkgName(stackLevel int) string {
	// todo try cach
	pc, _, _, _ := runtime.Caller(stackLevel)
	functionObject := runtime.FuncForPC(pc)
	arr := strings.Split(functionObject.Name(), ".")
	sPkg := strings.Split(arr[0], "/")
	return sPkg[len(sPkg)-1]
}

// getModuleName ...
func getModuleName(stackLevel int) string {
	// todo try cach
	_, modulePathName, _, _ := runtime.Caller(stackLevel)
	sModule := strings.Split(modulePathName, "/")
	return sModule[len(sModule)-1]
}

// getFuncName ...
func getFuncName(stackLevel int) string {
	// todo try cach
	pc, _, _, _ := runtime.Caller(stackLevel)
	functionObject := runtime.FuncForPC(pc)
	arr := strings.Split(functionObject.Name(), ".")
	return arr[len(arr)-1]
}

// getLine ...
func getLine(stackLevel int) string {
	// todo try cach
	_, _, line, _ := runtime.Caller(stackLevel)
	return strconv.Itoa(line)
}

// msgToJSON ...
func msgToJSON(m Msg) (string, error) {
	defer selfRecover()
	jsonMsg, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err) // todo format
		return "", err
	}
	return string(jsonMsg), nil
}

// TRACE ...
func TRACE(logText, logContext string) {
	defer selfRecover()
	sendMessage("TRACE", settings.StackLevelTrace, logContext, logText, "")
}

// INFO ...
func INFO(logText, logContext string) {
	defer selfRecover()
	sendMessage("INFO", settings.StackLevelInfo, logContext, logText, "")
}

// WARN ...
func WARN(logText, logContext string) {
	defer selfRecover()
	sendMessage("WARN", settings.StackLevelWarn, logContext, logText, "")
}

// ERROR ...
func ERROR(logText, logContext, errCode string) {
	defer selfRecover()
	sendMessage("ERROR", settings.StackLevelError, logContext, logText, errCode)
}