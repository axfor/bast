module github.com/aixiaoxiang/bast

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/aixiaoxiang/daemon v0.11.1
	github.com/julienschmidt/httprouter v1.2.0
	github.com/microsoft/go-winio v0.4.12
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190302025703-b6889370fb10

go 1.13
