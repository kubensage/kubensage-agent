package cli

import (
	"flag"
	"go.uber.org/zap"
	"time"
)

type AgentConfig struct {
	RelayAddress            string
	MainLoopDurationSeconds time.Duration
	BufferRetention         time.Duration
	TopN                    int
}

func RegisterAgentFlags(fs *flag.FlagSet) func(logger *zap.Logger) *AgentConfig {
	relayAddress := fs.String("relay-address", "", "Relay address (required)")
	mainLoopDuration := fs.Int("main-loop-duration", 5, "Main loop duration in seconds")
	bufferRetention := fs.Int("buffer-retention", 10, "Buffer retention in minutes")
	topN := fs.Int("top-n", 10, "Top N processes")

	return func(logger *zap.Logger) *AgentConfig {
		if *relayAddress == "" {
			logger.Fatal("Missing required flag: --relay-address")
		}

		return &AgentConfig{
			RelayAddress:            *relayAddress,
			MainLoopDurationSeconds: time.Duration(*mainLoopDuration) * time.Second,
			BufferRetention:         time.Duration(*bufferRetention) * time.Minute,
			TopN:                    *topN,
		}
	}
}
