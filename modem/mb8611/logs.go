package mb8611

import (
	"encoding/json"
	"strings"
)

type Logs struct {
	SOAPAction string
	Request    struct {
		GetMultipleHNAPs struct {
			GetMotoStatusLog    string `json:"GetMotoStatusLog"`
			GetMotoStatusLogXXX string `json:"GetMotoStatusLogXXX"`
		} `json:"GetMultipleHNAPs"`
	} `json:"-"`
	Response struct {
		GetMultipleHNAPsResponse struct {
			GetMotoStatusLogResponse struct {
				MotoStatusLogList      string `json:"MotoStatusLogList,omitempty"`
				GetMotoStatusLogResult string `json:"GetMotoStatusLogResult,omitempty"`
			} `json:"GetMotoStatusLogResponse,omitempty"`
			GetMotoStatusLogXXXResponse struct {
				XXX                       string `json:"XXX,omitempty"`
				GetMotoStatusLogXXXResult string `json:"GetMotoStatusLogXXXResult,omitempty"`
			} `json:"GetMotoStatusLogXXXResponse,omitempty"`
			GetMultipleHNAPsResult string `json:"GetMultipleHNAPsResult,omitempty"`
		} `json:"GetMultipleHNAPsResponse,omitempty"`
	}
}

func NewLogs() *Logs {
	var req Logs = Logs{}
	req.SOAPAction = "http://purenetworks.com/HNAP1/GetMultipleHNAPs"
	req.Request.GetMultipleHNAPs.GetMotoStatusLog = ""
	req.Request.GetMultipleHNAPs.GetMotoStatusLogXXX = ""
	return &req
}

func (l *Logs) RawLogMessages() string {
	return l.Response.GetMultipleHNAPsResponse.GetMotoStatusLogResponse.MotoStatusLogList
}

func (l *Logs) LogMessages() string {
	type rule struct {
		find    string
		replace string
	}
	parseRules := []rule{
		{find: "\n^", replace: " "},
		{find: "}-{", replace: "\n"},
		{find: "^", replace: " "},
	}

	ret := l.RawLogMessages()
	for _, r := range parseRules {
		ret = strings.ReplaceAll(ret, r.find, r.replace)
	}
	return ret
}

func (l *Logs) Action() string {
	return l.SOAPAction
}

func (l *Logs) Marshal() []byte {
	ret, _ := json.Marshal(l)
	return ret
}

func (l *Logs) MarshalRequest() []byte {
	ret, _ := json.Marshal(l.Request)
	return ret
}

func (l *Logs) MarshalIndent() []byte {
	ret, _ := json.MarshalIndent(l, "", "  ")
	return ret
}
