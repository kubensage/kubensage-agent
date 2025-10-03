package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kubensage/kubensage-agent/pkg/buildinfo"
	"go.uber.org/zap"
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
// This function defines all CLI flags on the provided FlagSet and returns a closure
// that, when executed, parses and validates their values into an *AgentConfig.
//
// The closure ensures required flags are provided, converts integer durations into
// proper time.Duration values, and optionally prints version information if requested.
//
// Required flag:
//
//	--relay-address string
//	  The address of the metrics relay gRPC server (e.g. "localhost:5000").
//
// Optional flags:
//
//	--main-loop-duration int
//	  Duration of the main collection loop in seconds (default: 5)
//
//	--buffer-retention int
//	  Total retention time in minutes for buffered metrics (default: 10)
//
//	--top-n int
//	  Number of top memory-consuming processes to report (default: 10)
//
//	--version
//	  If set, prints the current agent version (as defined in pkg/buildinfo.Version) and exits.
//
// Parameters:
//   - fs: *flag.FlagSet
//     The flag set to which the flags will be bound (usually flag.CommandLine).
//
// Returns:
//   - func(logger *zap.Logger) *AgentConfig
//     A closure that builds and returns a validated *AgentConfig.
//     If the --relay-address flag is missing, the closure will call logger.Fatal and terminate.
//     If --version is set, the closure prints the version string and exits with code 0.
func RegisterAgentFlags(
	fs *flag.FlagSet,
) func(logger *zap.Logger) *AgentConfig {
	relayAddress := fs.String("relay-address", "", "Relay address (required)")
	mainLoopDuration := fs.Int("main-loop-duration", 5, "Main loop duration in seconds")
	bufferRetention := fs.Int("buffer-retention", 10, "Buffer retention in minutes")
	topN := fs.Int("top-n", 10, "Top N processes")
	version := fs.Bool("version", false, "Print the current version and exit")

	return func(logger *zap.Logger) *AgentConfig {
		// Handle version flag
		if *version {
			fmt.Printf("%s\n", buildinfo.Version)
			os.Exit(0)
		}

		// Validate required flags
		if *relayAddress == "" {
			logger.Fatal("missing required flag: --relay-address")
		}

		// Build and return configuration
		return &AgentConfig{
			RelayAddress:            *relayAddress,
			MainLoopDurationSeconds: time.Duration(*mainLoopDuration) * time.Second,
			BufferRetention:         time.Duration(*bufferRetention) * time.Minute,
			TopN:                    *topN,
		}
	}
}
