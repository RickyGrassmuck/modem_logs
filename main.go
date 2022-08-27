package main

import (
	"fmt"
	"log"
	"os"
	"path"

	modem "github.com/RickyGrassmuck/modem_logs/modem/mb8611"
	"github.com/RickyGrassmuck/modem_logs/utils"
)

var logger *log.Logger
var defaultLogDir string
var defaultModemAddr string
var defaultLogFileName string = "modem_logs.txt"

func init() {
	defaultLogDir, _ = os.Getwd()
	defaultModemAddr = "https://192.168.100.1/HNAP1/"
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

type Config struct {
	ModemConfig *modem.ModemConfig
	LogFile     string
	ModemAddr   string
}

func setup() *Config {
	var err error
	conf := &Config{}

	username, ok := os.LookupEnv("MODEM_USERNAME")
	if !ok {
		logger.Println("Must set MODEM_USERNAME env variable")
		os.Exit(1)
	}
	password, ok := os.LookupEnv("MODEM_PASSWORD")
	if !ok {
		logger.Println("Must set MODEM_USERNAME env variable")
		os.Exit(1)
	}

	conf.LogFile, ok = os.LookupEnv("MODEM_LOG_DESTINATION")
	if !ok {
		logger.Printf("MODEM_LOG_DESTINATION not set, log will be saved to %s\n", path.Join(defaultLogDir, defaultLogFileName))
		conf.LogFile = defaultLogDir
	} else {
		logger.Printf("Log will be saved to %s\n", conf.LogFile)
	}

	conf.ModemAddr, ok = os.LookupEnv("MODEM_ADDRESS")
	if !ok {
		logger.Printf("MODEM_LOG_DESTINATION not set, log will be saved to %s\n", defaultModemAddr)
		conf.ModemAddr = defaultModemAddr
	}

	conf.ModemConfig, err = modem.NewClient(conf.ModemAddr)
	if err != nil {
		logger.Fatal(err)
	}

	auth, err := conf.ModemConfig.Login(username, password)
	if err != nil {
		logger.Printf("%v\n", err)
		os.Exit(1)
	}
	if auth.LoginResponse.LoginResult != "OK" {
		logger.Printf("Login failed: %v\n", auth.LoginResponse)
		os.Exit(1)
	}
	return conf
}

func appendFile(filepath string, data string) error {
	lastWriteFilePath := fmt.Sprintf("%s.%s", filepath, "last")
	tmpFilePath := fmt.Sprintf("%s.%s", filepath, "tmp")
	tmpFile, _ := os.OpenFile(tmpFilePath, os.O_RDWR|os.O_CREATE, 0644)
	defer tmpFile.Close()
	_, _ = tmpFile.WriteString(data)

	// Last write file exists, check if current data is different from last write file
	if _, err := os.Stat(lastWriteFilePath); err == nil {
		sameAsLast, _ := utils.DeepCompare(lastWriteFilePath, tmpFilePath)
		if sameAsLast {
			logger.Println("No new log messages, skipping...")
			os.Remove(tmpFilePath)
			return nil
		}
	} else {
		logger.Println("No last write file found, creating...")
		lastWriteFile, err := os.OpenFile(lastWriteFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer lastWriteFile.Close()
		lastWriteFile.WriteString(data)
	}
	aggregateLogFile, _ := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer aggregateLogFile.Close()
	_, err := aggregateLogFile.WriteString(data)
	os.Remove(tmpFilePath)
	return err
}

func main() {
	conf := setup()
	client := conf.ModemConfig
	logs, err := client.GetLogs()
	if err != nil {
		logger.Printf("%v\n", err)
		os.Exit(1)
	}
	err = appendFile(path.Join(conf.LogFile, defaultLogFileName), logs.LogMessages())
	if err != nil {
		logger.Printf("%v\n", err)
		os.Exit(1)
	}
}
