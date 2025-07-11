package utils

import (
	"flag"
	"log"
	"time"
)

type Config struct {
	RelayAddress            string
	MainLoopDurationSeconds time.Duration

	LogLevel      string
	LogFile       string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
	LogCompress   bool
}

func ParseFlags() Config {
	relayAddress := flag.String("relay-address", "",
		"The address of the relay grpc server, (Required: yes, Default: N/A)")

	mainLoopDurationSecondsFlag := flag.Int("main-loop-duration-seconds", 5,
		"The duration of the main loop (Required: No, Default: 5s)")

	logLevel := flag.String("log-level", "info",
		"Set log level, (Required: No, Default: info)")

	logFile := flag.String("log-file", "/var/log/kubensage/kubensage-agent.log",
		"Path to log file, (Required: No, Default: /var/log/kubensage-agent.log)")

	logMaxSize := flag.Int("log-max-size", 10,
		"Maximum log size (MB), (Required: No, Default: 10)")

	logMaxBackups := flag.Int("log-max-backups", 5,
		"Max backup files, (Required: No, Default: 5)")

	logMaxAge := flag.Int("log-max-age", 30,
		"Max age in days to retain old log files, (Required: No, Default: 30)")

	logCompress := flag.Bool("log-compress", true,
		"Compress old log files, (Required: No, Default: true)")

	flag.Parse()

	mainLoopDuration := time.Duration(*mainLoopDurationSecondsFlag) * time.Second

	if *relayAddress == "" {
		log.Fatal("Missing required flag: --relay-address")
	}

	return Config{
		RelayAddress:            *relayAddress,
		MainLoopDurationSeconds: mainLoopDuration,
		LogLevel:                *logLevel,
		LogFile:                 *logFile,
		LogMaxSize:              *logMaxSize,
		LogMaxBackups:           *logMaxBackups,
		LogMaxAge:               *logMaxAge,
		LogCompress:             *logCompress,
	}
}
