module github.com/kubensage/kubensage-agent

go 1.24.4

// replace github.com/kubensage/go-common => /home/roman/github/kubensage/go-common

require (
	github.com/kubensage/go-common v1.0.9
	github.com/shirou/gopsutil/v3 v3.24.5
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	k8s.io/cri-api v0.34.1
)

require (
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250827001030-24949be3fa54 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shoenig/go-m1cpu v0.1.7 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250922171735-9219d122eba9 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
