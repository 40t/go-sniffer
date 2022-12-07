package model

import "encoding/json"

// MysqlQueryPiece 查询信息
type MysqlQueryPiece struct {
	BaseQueryPiece

	ClientHost string `json:"cip"`
	ClientPort string `json:"cport"`

	VisitUser    string `json:"user"`
	VisitDB      string `json:"db"`
	QuerySQL     string `json:"sql"`
	CostTimeInMS int64  `json:"cms"`
	Message      string `json:"message"`
}

// BaseQueryPiece 查询信息
type BaseQueryPiece struct {
	ServerIP          string  `json:"sip"`
	ServerPort        string  `json:"sport"`
	CapturePacketRate float64 `json:"cpr"`
	EventTime         int64   `json:"bt"`
}

func (p MysqlQueryPiece) ToString() string {
	bytes, err := json.Marshal(&p)
	if err != nil {
		return ""
	}

	return string(bytes)
}
