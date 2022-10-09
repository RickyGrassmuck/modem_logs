package main

import (
	// "encoding/json"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	modem "github.com/RickyGrassmuck/modem_logs/modem/mb8611"
	"github.com/RickyGrassmuck/modem_logs/utils"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/shopspring/decimal"
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
	DebugMode   bool
	ModemConfig *modem.ModemConfig
	LogFile     string
	ModemAddr   string
	Influx      InfluxConfig
}

type InfluxConfig struct {
	URL    string
	Token  string
	Bucket string
	Org    string
	Client influxdb2.Client
}
type InfluxDownstream struct {
	Time              time.Time
	Channel           decimal.Decimal
	Power             decimal.Decimal
	SNR               decimal.Decimal
	CorrectedErrors   decimal.Decimal
	UncorrectedErrors decimal.Decimal
}

func getEnvOrExit(varName string) string {
	envVar, ok := os.LookupEnv(varName)
	if !ok {
		logger.Printf("Must set %s env variable", varName)
		os.Exit(1)
	}
	return envVar
}

func getEnvOrDefault(varName string, defaultVal string) string {
	envVar, ok := os.LookupEnv(varName)
	if !ok {
		logger.Printf("%s not set, using default: %s\n", varName, defaultVal)
		envVar = defaultVal
	}
	return envVar
}

func setup() *Config {
	var err error
	conf := &Config{Influx: InfluxConfig{}}
	_, ok := os.LookupEnv("MODEM_DEBUG")
	if !ok {
		conf.DebugMode = false
	} else {
		conf.DebugMode = true
	}

	username := getEnvOrExit("MODEM_USERNAME")
	password := getEnvOrExit("MODEM_PASSWORD")
	conf.ModemAddr = getEnvOrDefault("MODEM_ADDRESS", defaultModemAddr)
	conf.LogFile = getEnvOrDefault("MODEM_LOG_DESTINATION", defaultLogFileName)
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

	influxURL := getEnvOrExit("INFLUX_URL")
	influxToken := getEnvOrExit("INFLUX_TOKEN")
	influxBucket := getEnvOrExit("INFLUX_BUCKET")
	influxOrg := getEnvOrExit("INFLUX_ORG")

	conf.Influx = InfluxConfig{
		URL:    influxURL,
		Token:  influxToken,
		Bucket: influxBucket,
		Org:    influxOrg,
	}
	client := influxdb2.NewClient(conf.Influx.URL, conf.Influx.Token)

	conf.Influx.Client = client

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

func generateTable(title string, headers table.Row, records [][]string, output io.Writer) table.Writer {
	newTable := table.NewWriter()
	newTable.SetTitle("%s", title)
	newTable.SetStyle(table.StyleLight)
	newTable.Style().Title.Align = text.AlignCenter
	newTable.SetOutputMirror(os.Stdout)
	newTable.AppendHeader(headers)
	var rows []table.Row
	for _, v := range records[1:] {
		row := table.Row{}
		for _, r := range v {
			row = append(row, r)
		}
		rows = append(rows, row)

	}
	newTable.AppendRows(rows)

	return newTable
}

func (c *Config) saveLogs() {
	logs, err := c.ModemConfig.GetLogs()
	if err != nil {
		logger.Printf("%v\n", err)
		os.Exit(1)
	}
	err = appendFile(path.Join(c.LogFile, defaultLogFileName), logs.LogMessages())
	if err != nil {
		logger.Printf("%v\n", err)
		os.Exit(1)
	}
}

func printConnectionDetails(connDetails *modem.Connection) {

	dsTable := generateTable("DOWNSTREAM", modem.DownstreamHeaders, connDetails.Downstream.ToCSV(), os.Stdout)
	usTable := generateTable("UPSTREAM", modem.UpstreamHeaders, connDetails.Upstream.ToCSV(), os.Stdout)
	dsTable.Render()
	usTable.Render()
}

func downstreamStatsToInflux(connDetails modem.ConnectionDetails) []*write.Point {

	var points []*write.Point

	records := connDetails.ToCSV()
	influxTimeStamp := time.Now().UTC()

	powerLevels := []float64{}
	totalCorrected := decimal.NewFromInt(0)
	totalUncorrected := decimal.NewFromInt(0)

	for _, record := range records {
		channel := record[0]
		power, _ := decimal.NewFromString(record[5])
		snr, _ := decimal.NewFromString(record[6])
		correctedErrors, _ := decimal.NewFromString(record[7])
		uncorrectedErrors, _ := decimal.NewFromString(record[8])

		p := influxdb2.NewPointWithMeasurement("downstream").
			AddTag("id", channel).
			AddField("power", power).
			AddField("snr", snr).
			AddField("corrected_errors", correctedErrors).
			AddField("uncorrected_errors", uncorrectedErrors).
			SetTime(influxTimeStamp)

		totalCorrected = totalCorrected.Add(correctedErrors)
		totalUncorrected = totalUncorrected.Add(uncorrectedErrors)
		powerLevels = append(powerLevels, power.InexactFloat64())
		points = append(points, p)
	}

	sumPoint := influxdb2.NewPointWithMeasurement("downstream").
		AddTag("id", "101").
		AddField("power", decimal.NewFromFloat(0.0)).
		AddField("snr", decimal.NewFromFloat(0.0)).
		AddField("corrected_errors", totalCorrected).
		AddField("uncorrected_errors", totalUncorrected).
		SetTime(influxTimeStamp)

	logger.Printf("Total Corrected: %v\n", totalCorrected)
	logger.Printf("Total Uncorrected: %v\n", totalUncorrected)
	points = append(points, sumPoint)

	powerSpreadPoint := influxdb2.NewPointWithMeasurement("downstream").
		AddTag("id", "102").
		AddField("power", decimal.NewFromFloat(utils.CalculateSpread(powerLevels)).Round(1)).
		AddField("snr", decimal.NewFromFloat(0.0)).
		AddField("corrected_errors", decimal.NewFromFloat(0.0)).
		AddField("uncorrected_errors", decimal.NewFromFloat(0.0)).
		SetTime(influxTimeStamp)

	points = append(points, powerSpreadPoint)

	return points
}

func (c *Config) writeConnectionStatsInfluxdb(stats *modem.Connection) error {
	ctx := context.Background()
	bucketsAPI := c.Influx.Client.BucketsAPI()
	_, err := bucketsAPI.FindBucketByName(ctx, "modem_stats")
	if err != nil {
		logger.Printf("Bucket not found, creating...\n")
		org, err := c.Influx.Client.OrganizationsAPI().FindOrganizationByName(ctx, c.Influx.Org)
		if err != nil {
			panic(err)
		}
		_, err = bucketsAPI.CreateBucketWithName(ctx, org, c.Influx.Bucket, domain.RetentionRule{EverySeconds: 0})
		if err != nil {
			panic(err)
		}
		logger.Printf("Bucket Created")
	}
	writeAPI := c.Influx.Client.WriteAPI(c.Influx.Org, c.Influx.Bucket)
	dsPoints := downstreamStatsToInflux(stats.Downstream)
	for _, p := range dsPoints {
		writeAPI.WritePoint(p)
	}
	writeAPI.Flush()
	return nil
}

func main() {
	conf := setup()
	defer conf.Influx.Client.Close()

	for {
		connDetails, err := conf.ModemConfig.GetConnectionDetails()
		if err != nil {
			logger.Printf("%v\n", err)
		}
		conf.writeConnectionStatsInfluxdb(connDetails)
		fmt.Printf("Sleeping for 10 seconds...\n\n")
		time.Sleep(10 * time.Second)
	}

}
