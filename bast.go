//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	sdaemon "github.com/aixiaoxiang/daemon"
	"github.com/julienschmidt/httprouter"
)

var (
	usageline = `help:
	-h | -help                    show help
	-develop                      run in develop(develop environment)
	-start                        run in background
	-stop                         graceful for stop
	-reload                       graceful for reload
	-migration                    migration or initial system
	-conf=your path/config.conf   config path(default is ./config.conf)
	-install                      install to service
	-uninstall                    uninstall for service
	`
	flagDevelop, flagStart, flagStop, flagReload, flagDaemon                     bool
	isInstall, isUninstall, isForce, flagService, isMaster, isClear, isMigration bool
	flagConf, flagName, flagAppKey, flagPipe                                     string
	flagPPid                                                                     int
	app                                                                          *App
	t, f                                                                         = true, false
	//T *bool
	T = &t
	//F *bool
	F = &f
)

//App is application major data
type App struct {
	pool                                 sync.Pool
	Router                               *httprouter.Router
	Addr, pipeName                       string
	Server                               *http.Server
	Before                               BeforeHandle
	After                                AfterHandle
	Migration                            MigrationHandle
	Debug, Daemon, isCallCommand, runing bool
	cmd                                  []work
}

type work struct {
	key       string
	cmd       *exec.Cmd
	runing    bool
	exitCount int
}

//init application
func init() {
	os.Chdir(AppDir())
	app = &App{Server: &http.Server{}, Router: httprouter.New(), runing: true}
	parseCommandLine()
	app.pool.New = func() interface{} {
		return &Context{}
	}
	//init config
	if ConfOK() {
		logs.LogInit(LogConf())
	}
	//register http OPTIONS router
	doHandle("OPTIONS", "/*filepath", nil)
}

//parseCommandLine parse commandLine
func parseCommandLine() {
	f := flag.NewFlagSet("bast", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(usageline)
		os.Exit(0)
	}
	if len(os.Args) == 2 && (os.Args[1] == "h" || os.Args[1] == "help") {
		f.Usage()
	}
	f.BoolVar(&flagDevelop, "develop", false, "")
	f.BoolVar(&flagStart, "start", false, "")
	f.BoolVar(&flagStop, "stop", false, "")
	f.BoolVar(&flagReload, "reload", false, "")
	f.BoolVar(&flagDaemon, "daemon", false, "")
	f.BoolVar(&isUninstall, "uninstall", false, "")
	f.BoolVar(&isForce, "force", false, "")
	f.BoolVar(&isInstall, "install", false, "")
	f.BoolVar(&flagService, "service", false, "")
	f.BoolVar(&isMaster, "master", false, "")
	f.BoolVar(&isMigration, "migration", false, "")
	f.StringVar(&flagName, "name", "", "")
	f.StringVar(&flagConf, "conf", "", "")
	f.StringVar(&flagAppKey, "appkey", "", "")
	f.StringVar(&flagPipe, "pipe", "", "")
	f.IntVar(&flagPPid, "pid", 0, "")

	f.Parse(os.Args[1:])
	if len(os.Args) == 1 {
		//flagStart = true
	}
	if flagName != "" {
		isInstall = true
	}
	if !isInstall {
		for _, k := range os.Args[1:] {
			if strings.HasPrefix(k, "-install") {
				isInstall = true
				break
			}
		}
	}
	if isInstall {
		flagDaemon = false
	}
	if flagDevelop || flagStop || flagReload || flagDaemon || isInstall || isUninstall || flagService {
		flagStart = false
	}
	if flagService {
		isMaster = flagService
	}
	confParse(f)
}

//BeforeHandle is before then request handler
type BeforeHandle func(ctx *Context) error

//AfterHandle is after then request handler
type AfterHandle func(ctx *Context) error

//MigrationHandle is migration handler
type MigrationHandle func() error

//Before the request
func Before(f BeforeHandle) {
	app.Before = f
}

//After the request
func After(f AfterHandle) {
	app.After = f
}

//Migration set migration handler
func Migration(f MigrationHandle) {
	app.Migration = f
}

// ListenAndServe see net/http ListenAndServe
func (app *App) ListenAndServe() error {
	app.Server.Addr = app.Addr
	app.Server.Handler = app.Router
	return app.Server.ListenAndServe()
}

// Post registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Post(pattern string, f func(ctx *Context), authorization ...bool) {
	doHandle("POST", pattern, f, authorization...)
}

// Get registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Get(pattern string, f func(ctx *Context), authorization ...bool) {
	doHandle("GET", pattern, f, authorization...)
}

// FileServer registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func FileServer(pattern string, root string) {
	app.Router.Handler("GET", pattern+"*filepath", NoLookDirHandler(http.StripPrefix(pattern, http.FileServer(http.Dir(root)))))
}

//NoLookDirHandler disable directory look
func NoLookDirHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		if strings.Index(r.URL.RawQuery, "download") >= 0 {
			w.Header().Add("Content-Type", "application/octet-stream")
			r.ParseForm()
			fn := r.Form["rawName"]
			if fn != nil {
				w.Header().Add("Content-Type", "application/octet-stream")
				w.Header().Add("Content-Disposition", "attachment;filename=\""+fn[0]+"\"")
			}
		}
		h.ServeHTTP(w, r)
	})
}

// doHandle registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func doHandle(method, pattern string, f func(ctx *Context), authorization ...bool) {
	auth := false
	if authorization != nil {
		auth = authorization[0]
	}
	//app.Router.HandlerFunc(method,pattern)
	app.Router.Handle(method, pattern, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logs.Info(r.Method + ":" + r.RequestURI + "->start")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization,Access-Control-Allow-Origin,Content-Length,Content-Type,BaseUrl")
			w.Header().Set("Access-Control-Max-Age", "1728000")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == "OPTIONS" {
			return
		}
		if pattern == "/" && r.URL.Path != pattern {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, http.StatusText(http.StatusNotFound))
			return
		}
		if r.Method == method {
			if f != nil {
				ctx := app.pool.Get().(*Context)
				ctx.Reset()
				defer app.pool.Put(ctx)
				ctx.In = r
				ctx.Out = w
				ctx.Params = ps
				ctx.Authorization = auth
				defer func() {
					if err := recover(); err != nil {
						errMsg := fmt.Sprintf("%s", err)
						logs.Error(r.Method + ":" + r.RequestURI + "->error=" + errMsg)
						w.WriteHeader(http.StatusInternalServerError)
						fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
					}
				}()
				if app.Before != nil {
					if app.Before(ctx) != nil {
						logs.Info(r.Method + ":" + r.RequestURI + "->end")
						return
					}
				}
				f(ctx)
				if app.After != nil {
					app.After(ctx)
				}
				logs.Info(r.Method + ":" + r.RequestURI + "->end")
			}
		} else {
			logs.Info(r.Method + ":" + r.RequestURI + "->end=Not allowed")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, http.StatusText(http.StatusMethodNotAllowed))
		}
	})
}

//Run use addr to start app
func Run(addr string) {
	if !app.isCallCommand && !Command() {
		return
	}
	defer clear()
	doRun(addr)
}

//Serve use config to start app
func Serve() bool {
	c := Conf()
	if !app.isCallCommand && !Command() || c == nil {
		return false
	}
	defer clear()

	Debug(c.Debug)
	Run(c.Addr)
	return true
}

// TryServe try to check the configuration can be turned on
// 1: is control commandline
// 2: config is ok
func TryServe() bool {
	ok := Command() && ConfOK()
	// defer func() {
	// 	fmt.Printf("finish=%v\r\n", ok)
	// }()
	return ok
}

//Debug set app debug status
func Debug(debug bool) {
	app.Debug = debug
}

//doRun real run app
func doRun(addr string) {
	app.Addr = addr
	err := tryRun()
	if err == nil {
		logs.Info("addr=" + app.Addr)
		// fmt.Println("start")
		err = app.ListenAndServe()
		if err != nil {
			fmt.Println("listenAndServe error=" + err.Error())
			logs.Info("listenAndServe error=" + err.Error())
			os.Exit(-1)
		}
		// fmt.Println("finish")
		logs.Info("finish")
	} else {
		// fmt.Println("listen error=" + err.Error())
		logs.Info("listen error=" + err.Error())
		os.Exit(-1)
	}
}

func tryRun() error {
	l, err := net.Listen("tcp", app.Addr)
	if err != nil {
		return err
	}
	l.Close()
	l = nil
	return nil
}

//Command Commandline args
func Command() bool {
	if app.isCallCommand {
		return false
	}
	app.isCallCommand = true
	r := true
	var err error
	if isMigration {
		if app.Migration != nil {
			err := app.Migration()
			if err != nil {
				fmt.Println("migration error,detail:" + err.Error())
				return false
			}
		}
	}
	if flagStart {
		r, err = start()
	} else if flagService {
		service()
		// err = errors.New("service child process")
		r = false
	} else if flagStop {
		stop()
		// err = errors.New("stop child process")
		r = false
	} else if flagReload {
		reload()
		// err = errors.New("inside child process for reload")
		r = false
	} else if flagDaemon {
		daemon()
	} else if isInstall {
		// err = errors.New("install service")
		install()
		r = false
	} else if isUninstall {
		// err = errors.New("uninstall service")
		uninstall()
		r = false
	}
	if err != nil {
		// fmt.Println(err.Error())
	}
	return r
}

func start() (bool, error) {
	if isMaster {
		doStart()
		checkWorkProcess()
		return false, nil
	}
	path := ConfPath()
	cmd := exec.Command(os.Args[0], "-master", "-start", "-conf="+path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = AppDir()
	cmd.Start()
	time.Sleep(200 * time.Microsecond)
	// os.Exit(0)
	return false, nil
}

func service() {
	if flagName == "" {
		flagName = AppName()
	}
	service, err := sdaemon.New(flagName, flagName+" service")
	if err != nil {
		logs.Info("service failed," + err.Error())
		return
	}
	status, err := service.Run(&daemonExecutable{})
	if err != nil {
		logs.Info("service failed," + err.Error())
		return
	}
	logs.Info("service info:" + status)

}

func doService() {
	doStart()
	go checkWorkProcess()
}

func doStart() error {
	appConfs := Confs()
	path := ConfPath()
	pid := strconv.Itoa(os.Getpid())
	app.pipeName = pid
	if flagService {
		logs.Info("service=" + path + ",master pid=" + pid)
	} else {
		// fmt.Println("start=" + path + ",master pid=" + pid)
	}
	app.cmd = []work{}
	for _, c := range appConfs {
		cmd := exec.Command(os.Args[0], "-daemon", "-appkey="+c.Key, "-pipe="+app.pipeName, "-conf="+path)
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		cmd.Dir = AppDir()
		cmd.Start()
		app.cmd = append(app.cmd, work{key: c.Key, cmd: cmd, runing: true})
	}
	if err := logPid(); err != nil {
		logs.Err("start error log pid,", err)
		if !flagService {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func startWork(index int) *exec.Cmd {
	w := app.cmd[index]
	c := ConfWithKey(w.key)
	if c != nil {
		path := ConfPath()
		// pid := strconv.Itoa(os.Getpid())
		cmd := exec.Command(os.Args[0], "-daemon", "-appkey="+c.Key, "-pipe="+app.pipeName, "-conf="+path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = AppDir()
		cmd.Start()
		if app.cmd == nil {
			app.cmd = nil
		}
		app.cmd[index] = work{key: c.Key, cmd: cmd, runing: true}
		logPid()
		return cmd
	}
	return nil
}

//checkWorkProcess check work process stat
func checkWorkProcess() {
	c := make(chan struct{})
	lg := len(app.cmd)
	l := 0
	for i := 0; i < lg; i++ {
		w := app.cmd[i]
		if w.runing {
			go func(wc work) {
				w.cmd.Wait()
				c <- struct{}{}
			}(w)
			l++
		}
	}
	for i := 0; i < lg; i++ {
		<-c
		w := app.cmd[i]
		exitCode := ""
		if w.cmd.ProcessState != nil {
			exitCode = strconv.Itoa(w.cmd.ProcessState.ExitCode())
		}
		if flagService {
			logs.Error("has work process exited,exit code=" + exitCode)
		} else {
			fmt.Println("has work process exited,exit code=" + exitCode)
		}
		w.runing = false
		if app.runing {
		} else {
			break
		}
	}
	if flagService {
		logs.Error("exited check work process")
	} else {
		fmt.Println("exited check work process")
	}
	clear()
	// os.Exit(0)
}

type daemonExecutable struct {
}

func (e *daemonExecutable) Start() {

}

func (e *daemonExecutable) Stop() {
	app.runing = false
	serviceStop()
}

func (e *daemonExecutable) Run() {
	doService()
}

func reload() {
	pids := getWorkPids()
	for _, pid := range pids {
		sendSignal(syscall.SIGINT, pid)
	}
	time.Sleep(200 * time.Millisecond)
	start()
}

func serviceStop() {
	pids := getWorkPids()
	for _, pid := range pids {
		sendSignal(syscall.SIGINT, pid)
	}
	// time.Sleep(10 * time.Millisecond)
	clear()
}

func stop() {
	pids := getWorkPids()
	for _, pid := range pids {
		sendSignal(syscall.SIGINT, pid)
	}
	time.Sleep(200 * time.Millisecond)
	os.Exit(0)
}

//AppDir app path
func AppDir() string {
	exPath := filepath.Dir(os.Args[0])
	return exPath
}

//AppName  app name
func AppName() string {
	fn := path.Base(os.Args[0])
	fileSuffix := path.Ext(fn)
	fn = strings.TrimSuffix(fn, fileSuffix)
	return fn
}

func sendSignal(sig os.Signal, pid int) error {
	pro, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		//TO DO  use pipe
		pro.Kill()
	} else {
		err = pro.Signal(sig)
		if err != nil {
			return err
		}
	}
	return nil
}

func daemon() {
	app.Daemon = true
	go signalListen()
}

func install() {
	if flagName == "" {
		flagName = AppName()
	}
	var agrs = []string{"-service", "-force", "-conf=" + flagConf}

	service, err := sdaemon.New(flagName, flagName+" service")
	if err != nil {
		fmt.Println("install failed," + err.Error())
		return
	}
	_, err = service.Install(agrs...)
	if err != nil {
		fmt.Println("install failed," + err.Error())
		return
	}
	fmt.Println("install success")
}

func uninstall() {
	if flagName == "" {
		flagName = AppName()
	}
	service, err := sdaemon.New(flagName, flagName+" service")
	if err != nil {
		fmt.Println("uninstall failed," + err.Error())
		return
	}
	_, err = service.Remove()
	if err != nil {
		fmt.Println("uninstall failed," + err.Error())
		return
	}
	fmt.Println("uninstall success")
}

func signalListen() {
	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c)
	for {
		s := <-c
		if s == syscall.SIGINT || (runtime.GOOS == "windows" && s == os.Interrupt) {
			logs.Info("signal=" + s.String())
			signal.Stop(c)
			err := Shutdown(nil)
			if err != nil {
				logs.Info("shutdown-error=" + err.Error())
			} else {
				logs.Info("shutdown-success")
			}
			break
		}
	}
}

func logPid() error {
	pidPath := os.Args[0] + ".pid"
	f, err := os.OpenFile(pidPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errors.New("cannot be created pid file")
	}
	defer f.Close()
	pids := strconv.Itoa(os.Getpid()) + "|" + app.pipeName + ":"
	for index, p := range app.cmd {
		if p.runing {
			if index > 0 {
				pids += ","
			}
			pids += strconv.Itoa(p.cmd.Process.Pid)
		}
	}
	if _, err := f.Write([]byte(pids)); err != nil {
		return errors.New("cannot be write pid file")
	}
	f.Sync()
	return nil
}

func removePid() error {
	pidPath := os.Args[0] + ".pid"
	return os.Remove(pidPath)
}

//getWorkPids work pids
func getWorkPids() []int {
	pidPath := os.Args[0] + ".pid"
	f, _ := os.Open(pidPath)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil
	}
	c := string(data)
	cs := strings.Split(c, ":")
	if len(cs) < 2 {
		return nil
	}
	c = cs[1]
	cs = strings.Split(c, ",")
	pids := []int{}
	for _, v := range cs {
		vv, err := strconv.Atoi(v)
		if err == nil {
			pids = append(pids, vv)
		}
	}
	if len(pids) > 0 {
		return pids
	}
	return nil
}

//pidPath pid filename path
func pidPath(path ...string) string {
	pidPath := ""
	if path != nil {
		pidPath = path[0]
	}
	if pidPath == "" {
		pidPath = ConfPath() + ".pid"
	} else {
		pidPath += ".pid"
	}
	// fmt.Printf("%s\n", pidPath)
	return pidPath
}

//Shutdown app
func Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
		// var cf context.CancelFunc
		// ctx, cf = context.WithTimeout(context.Background(), 30*time.Second)
		// if cf != nil {
		// 	//
		// }
	}
	app.Server.SetKeepAlivesEnabled(false)
	return app.Server.Shutdown(ctx)
}

/******ID method **********/

//ID create Unique ID
func ID() int64 {
	return ids.ID()
}

/******GUID method **********/

//GUID create GUID
func GUID() string {
	return guid.GUID()
}

//clear res
func clear() {
	if isClear {
		return
	}
	isClear = true
	logs.ClearLogger()
	ids.IDClear()
	removePid()
}
