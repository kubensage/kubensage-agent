package cli

import (
	"flag"
	"go.uber.org/zap"
	"time"
)

// AgentConfig holds runtime configuration parameters for the agent,
// parsed from command-line flags.
type AgentConfig struct {
	RelayAddress            string        // Address of the relay gRPC server
	MainLoopDurationSeconds time.Duration // Duration of the main collection loop
	BufferRetention         time.Duration // Retention time for buffered metrics
	TopN                    int           // Number of top memory-consuming processes to track
}

// RegisterAgentFlags registers the CLI flags required to configure the kubensage agent.
//
// This function defines flags on the provided FlagSet and returns a closure,
// which, when executed, parses the values into an AgentConfig struct.
// The closure also validates required flags and converts durations into proper time.Duration values.
//
// Required flag:
//
//	--relay-address: string, the address of the metrics relay (e.g. "localhost:5000")
//
// Optional flags:
//
//	--main-loop-duration: int, duration of the metrics collection loop in seconds (default: 5)
//	--buffer-retention: int, total retention time in minutes for buffered metrics (default: 10)
//	--top-n: int, number of top memory-consuming processes to report (default: 10)
//
// Parameters:
//   - fs: *flag.FlagSet - the flag set to which the flags are bound (usually flag.CommandLine)
//
// Returns:
//   - func(logger *zap.Logger) *AgentConfig:
//     A closure that builds and returns a validated *AgentConfig.
//     If the --relay-address is missing, the closure will call logger.Fatal and terminate the program.
func RegisterAgentFlags(
	fs *flag.FlagSet,
) func(logger *zap.Logger) *AgentConfig {
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
			MainLoopDurationSeconds: time.Duration(*mainLoopDuration) * time.Millisecond,
			BufferRetention:         time.Duration(*bufferRetention) * time.Minute,
			TopN:                    *topN,
		}
	}
}
