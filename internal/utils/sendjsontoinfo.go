package utils

import (
	"encoding/json"
	"strings"

	"github.com/cisco-pxgrid/cloud-sdk-go/log"
)

// Take a byte array that contains JSON, decode it, prettify it and send it
// line-by-line to log.Infof
func SendJsonToInfo(logger *log.DefaultLogger, msg string, b []byte) {
	var decoded interface{}
	json.Unmarshal(b, &decoded)
	pretty, _ := json.MarshalIndent(decoded, "", "  ")
	logger.Infof("message=%s", msg)
	for _, line := range strings.Split(string(pretty), "\n") {
		logger.Infof("%s", line)
	}

}
