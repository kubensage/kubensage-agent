package cli

import (
	"flag"
	"go.uber.org/zap"
	"time"
)

type AgentConfig struct {
	RelayAddress            string
	MainLoopDurationSeconds time.Duration
	TopN                    int
}

func RegisterAgentFlags(fs *flag.FlagSet) func(logger *zap.Logger) *AgentConfig {
	relayAddress := fs.String("relay-address", "", "Relay address (required)")
	mainLoopDuration := fs.Int("main-loop-duration-seconds", 5, "Main loop duration in seconds")
	topN := fs.Int("top-n", 10, "Top N processes")

	return func(logger *zap.Logger) *AgentConfig {
		if *relayAddress == "" {
			logger.Fatal("Missing required flag: --relay-address")
		}

		return &AgentConfig{
			RelayAddress:            *relayAddress,
			MainLoopDurationSeconds: time.Duration(*mainLoopDuration) * time.Second,
			TopN:                    *topN,
		}
	}
}
