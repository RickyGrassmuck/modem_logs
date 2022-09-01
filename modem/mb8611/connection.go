package mb8611

import (
	"encoding/csv"
	"encoding/json"
	// "fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var UpstreamHeaders table.Row = table.Row{
	"Channel",
	"Lock Status",
	"Channel Type",
	"Channel ID",
	"Symb. Rate (Ksym/sec)",
	"Freq. (MHz)",
	"Pwr (dBmV)",
}

var DownstreamHeaders table.Row = table.Row{
	"Channel",
	"Lock Status",
	"Modulation",
	"Channel ID",
	"Freq. (MHz)",
	"Pwr (dBmV)",
	"SNR (dB)",
	"Corrected",
	"Uncorrected",
}

type ConnectionData struct {
	SOAPAction string
	Request    struct {
		GetMultipleHNAPs struct {
			StartupSequence       string `json:"GetMotoStatusStartupSequence"`
			ConnectionInfo        string `json:"GetMotoStatusConnectionInfo"`
			DownstreamChannelInfo string `json:"GetMotoStatusDownstreamChannelInfo"`
			UpstreamChannelInfo   string `json:"GetMotoStatusUpstreamChannelInfo"`
			GetMotoLagStatus      string `json:"GetMotoLagStatus"`
		} `json:"GetMultipleHNAPs"`
	} `json:"-"`
	Response struct {
		StartupSequence struct {
			DSFreq                   string `json:"MotoConnDSFreq"`
			DSComment                string `json:"MotoConnDSComment"`
			ConnectivityStatus       string `json:"MotoConnConnectivityStatus"`
			ConnectivityComment      string `json:"MotoConnConnectivityComment"`
			BootStatus               string `json:"MotoConnBootStatus"`
			BootComment              string `json:"MotoConnBootComment"`
			ConfigurationFileStatus  string `json:"MotoConnConfigurationFileStatus"`
			ConfigurationFileComment string `json:"MotoConnConfigurationFileComment"`
			SecurityStatus           string `json:"MotoConnSecurityStatus"`
			SecurityComment          string `json:"MotoConnSecurityComment"`
			Result                   string `json:"GetMotoStatusStartupSequenceResult"`
		} `json:"GetMotoStatusStartupSequenceResponse"`
		ConnectionInfo struct {
			SystemUpTime         string `json:"MotoConnSystemUpTime"`
			NetworkAccess        string `json:"MotoConnNetworkAccess"`
			ConnectionInfoResult string `json:"GetMotoStatusConnectionInfoResult"`
		} `json:"GetMotoStatusConnectionInfoResponse"`
		DownstreamChannelInfoResponse struct {
			DownstreamChannel           string `json:"MotoConnDownstreamChannel"`
			DownstreamChannelInfoResult string `json:"GetMotoStatusDownstreamChannelInfoResult"`
		} `json:"GetMotoStatusDownstreamChannelInfoResponse"`
		UpstreamChannelInfoResponse struct {
			UpstreamChannel           string `json:"MotoConnUpstreamChannel"`
			UpstreamChannelInfoResult string `json:"GetMotoStatusUpstreamChannelInfoResult"`
		} `json:"GetMotoStatusUpstreamChannelInfoResponse"`
		GetMotoLagStatusResponse struct {
			MotoLagCurrentStatus   string `json:"MotoLagCurrentStatus"`
			GetMotoLagStatusResult string `json:"GetMotoLagStatusResult"`
		} `json:"GetMotoLagStatusResponse"`
		GetMultipleHNAPsResult string `json:"GetMultipleHNAPsResult"`
	} `json:"GetMultipleHNAPsResponse"`
}

type ConnectionDetails struct {
	Raw string
}

type Connection struct {
	ConnectivityStatus  string
	Uptime              string
	DownstreamFrequency string
	Upstream            ConnectionDetails
	Downstream          ConnectionDetails
}

func NewConnectionDetails() *ConnectionData {
	var conn ConnectionData = ConnectionData{}
	conn.SOAPAction = "http://purenetworks.com/HNAP1/GetMultipleHNAPs"
	conn.Request.GetMultipleHNAPs.StartupSequence = ""
	conn.Request.GetMultipleHNAPs.ConnectionInfo = ""
	conn.Request.GetMultipleHNAPs.DownstreamChannelInfo = ""
	conn.Request.GetMultipleHNAPs.UpstreamChannelInfo = ""
	conn.Request.GetMultipleHNAPs.GetMotoLagStatus = ""
	return &conn
}

func (c *ConnectionData) SanitizedDetails() *Connection {
	details := Connection{
		ConnectivityStatus:  c.Response.StartupSequence.ConnectivityStatus,
		Uptime:              c.Response.ConnectionInfo.SystemUpTime,
		DownstreamFrequency: c.Response.StartupSequence.DSFreq,
	}
	details.Upstream.Raw = c.Response.UpstreamChannelInfoResponse.UpstreamChannel
	details.Downstream.Raw = c.Response.DownstreamChannelInfoResponse.DownstreamChannel

	type rule struct {
		find    string
		replace string
	}
	parseRules := []rule{
		{find: "^|+|", replace: "\n"},
		{find: "^ ", replace: ","},
		{find: "^", replace: ","},
	}

	for _, r := range parseRules {
		details.Upstream.Raw = strings.ReplaceAll(details.Upstream.Raw, r.find, r.replace)
		details.Downstream.Raw = strings.ReplaceAll(details.Downstream.Raw, r.find, r.replace)
	}
	details.Upstream.Raw = strings.TrimSuffix(details.Upstream.Raw, ",")
	details.Downstream.Raw = strings.TrimSuffix(details.Downstream.Raw, ",")
	// fmt.Printf("%v", details)
	return &details
}

func (d *ConnectionDetails) ToCSV() [][]string {
	reader := csv.NewReader(strings.NewReader(d.Raw))
	records, _ := reader.ReadAll()
	return records
}

func (c *ConnectionData) Action() string {
	return c.SOAPAction
}

func (c *ConnectionData) Marshal() []byte {
	ret, _ := json.Marshal(c)
	return ret
}

func (c *ConnectionData) MarshalRequest() []byte {
	ret, _ := json.Marshal(c.Request)
	return ret
}

func (c *ConnectionData) MarshalIndent() []byte {
	ret, _ := json.MarshalIndent(c, "", "  ")
	return ret
}
