module github.com/aixiaoxiang/bast

go 1.12

require (
	github.com/aixiaoxiang/daemon v0.0.0-20190228090122-66ae0e7bb0f9
	github.com/julienschmidt/httprouter v1.2.0
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

// github.com/aixiaoxiang/daemon => /code/github/daemon
replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190228071610-92a0ff1e1e2f
