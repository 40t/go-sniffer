package build

import "time"

func GetNowStr(isClient bool) string {
	var msg string
	msg += time.Now().Format("2006-01-02 15:04:05.000")
	if isClient {
		msg += "| cli -> ser |"
	} else {
		msg += "| ser -> cli |"
	}
	return msg
}
