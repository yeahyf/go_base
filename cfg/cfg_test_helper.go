package cfg

import (
	"os"

	"github.com/yeahyf/go_base/log"
)

func initTestLogger() {
	logFile := "test_zap.json"
	testConfig := `{
		"level": "info",
		"logs": [
			{
				"logpath": "/tmp/test_debug.log",
				"maxsize": 100,
				"backups": 3,
				"maxage": 7,
				"compress": false,
				"name": "debug",
				"type": 0,
				"rotation": 1
			},
			{
				"logpath": "/tmp/test_info.log",
				"maxsize": 100,
				"backups": 3,
				"maxage": 7,
				"compress": false,
				"name": "info",
				"type": 0,
				"rotation": 1
			},
			{
				"logpath": "/tmp/test_error.log",
				"maxsize": 100,
				"backups": 3,
				"maxage": 7,
				"compress": false,
				"name": "error",
				"type": 0,
				"rotation": 1
			}
		]
	}`
	
	err := os.WriteFile(logFile, []byte(testConfig), 0644)
	if err != nil {
		panic(err)
	}
	defer os.Remove(logFile)
	
	log.SetLogConf(&logFile)
}
