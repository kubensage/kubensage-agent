module gitlab.com/kubensage/kubensage-agent

go 1.24.4

// replace gitlab.com/kubensage/go-common => /home/roman/gitlab/kubensage/go-common

require (
	github.com/shirou/gopsutil/v3 v3.24.5
	gitlab.com/kubensage/go-common v0.0.4
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
	k8s.io/cri-api v0.33.2
)

require (
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
