//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"context"
	"errors"
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

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/httpc"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/lang"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/aixiaoxiang/bast/session"
	"github.com/aixiaoxiang/daemon"

	"github.com/aixiaoxiang/bast/service"
	"github.com/aixiaoxiang/bast/snowflake"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

// var
var (
	flagDevelop, flagStart, flagStop, flagReload, flagDaemon                     bool
	isInstall, isUninstall, isForce, flagService, isMaster, isClear, isMigration bool
	flagConf, flagName, flagAppKey, flagPipe                                     string
	flagPid                                                                      int
	app                                                                          *App
	Log                                                                          *logs.Log
)

//App is application major data
type App struct {
	pool                                      sync.Pool
	Router                                    *httprouter.Router
	pattern                                   map[string]*Pattern
	registry                                  *service.Registry
	discovery                                 *service.Discovery
	Addr, pipeName, CertFile, KeyFile         string
	Server                                    *http.Server
	Before                                    BeforeHandle
	After                                     AfterHandle
	Authorization                             AuthorizationHandle
	Migration                                 MigrationHandle
	Debug, Daemon, isCallCommand, runing, tls bool
	cmd                                       []work
	cors                                      *conf.CORS
	wrap                                      bool
	id                                        *snowflake.Node
	page                                      *conf.Pagination
}

type work struct {
	key       string
	cmd       *exec.Cmd
	runing    bool
	exitCount int
}

//BeforeHandle a before handler for each request
type BeforeHandle func(ctx *Context) error

//AuthorizationHandle a authorization handler for each request
type AuthorizationHandle func(ctx *Context) error

//AfterHandle a after handler for each request
type AfterHandle func(ctx *Context) error

//MigrationHandle is migration handler
type MigrationHandle func() error

//init application
func init() {
	os.Chdir(AppDir())
	app = &App{Server: &http.Server{}, Router: httprouter.New(), runing: true, wrap: true, pattern: map[string]*Pattern{}}
	parseCommandLine()
	app.pool.New = func() interface{} {
		return &Context{}
	}
	//init module config
	if conf.OK() {
		Log = logs.Init(conf.LogConf())
		session.Init(conf.SessionConf())
		lang.TransFile(conf.Trans())
	} else {
		Log = logs.Init(nil)
	}
	app.cors = conf.CORSConf()

	app.wrap = conf.Wrap()

	app.id = ids.New()

	app.page = conf.PageConf()

	//register not found handler of router
	app.Router.NotFound = NotFoundHandler{}
	//register not allowed handler of router
	app.Router.MethodNotAllowed = MethodNotAllowedHandler{}
	//register options handler of router
	app.Router.GlobalOPTIONS = MethodOptionsHandler{}
}

//Before set the request 'before' handle
func Before(f BeforeHandle) {
	app.Before = f
}

//Auth set the request 'authorization' handle
func Auth(f AuthorizationHandle) {
	app.Authorization = f
}

//After  set the request 'after' handle
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

// ListenAndServeTLS see net/http ListenAndServeTLS
func (app *App) ListenAndServeTLS() error {
	app.Server.Addr = app.Addr
	app.Server.Handler = app.Router
	return app.Server.ListenAndServeTLS(app.CertFile, app.KeyFile)
}

// All registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func All(pattern string, f func(ctx *Context)) {
	for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions} {
		routerHandle(m, pattern, f)
	}
}

// Get registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Get(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodGet, pattern, f)
}

// Post registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Post(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodPost, pattern, f)
}

// Put registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Put(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodPut, pattern, f)
}

// Delete registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Delete(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodDelete, pattern, f)
}

// Head registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Head(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodHead, pattern, f)
}

// Patch registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Patch(pattern string, f func(ctx *Context)) *Pattern {
	return routerHandle(http.MethodPatch, pattern, f)
}

// Options registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Options(pattern string, f func(ctx *Context)) *Pattern {
	// doHandle(http.MethodOptions, pattern, f)
	return routerHandle(http.MethodOptions, pattern, f)
}

func routerHandle(method, pattern string, fn func(ctx *Context)) *Pattern {
	r := &Pattern{
		Method:  method,
		Pattern: pattern,
		Fn:      fn,
	}
	app.pattern[method+pattern] = r
	return r
}

//Router register to httpRouter
func Router() {
	for _, p := range app.pattern {
		pRef := p
		pRef.Router()
	}
	go Publish()
}

func initService() error {
	var err error
	c := conf.ServiceConf()
	if c != nil && c.Enable {
		err = initRegistry(c)
		if err != nil {
			return err
		}
		err = initDiscovery(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func initRegistry(c *conf.Service) error {
	var err error
	if app.registry == nil {
		app.registry, err = service.NewRegistry(c)
		if err != nil {
			logs.Errors("create registry failed", err)
			return err
		}
	}
	return nil
}

func initDiscovery(c *conf.Service) error {
	var err error
	if app.discovery == nil {
		app.discovery, err = service.NewDiscovery(c)
		if err != nil {
			logs.Errors("create discovery failed", err)
			return err
		}
		httpc.InitDiscovery(app.discovery)
	}
	return nil
}

//Publish publish to registry
func Publish() error {
	initService()
	var err error
	c := conf.ServiceConf()
	if c != nil && c.Enable && app.pattern != nil && app.registry != nil {
		pubs := 0
		ctx := context.Background()
		for _, p := range app.pattern {
			pRef := p
			if !pRef.publish {
				continue
			}
			_, err = app.registry.Put(ctx, c.Prefix+pRef.Service, c.Prefix+pRef.Pattern)
			if err != nil {
				continue
			}
			pRef.publishFinish = err == nil
			pubs++
		}
		if pubs > 0 {
			app.registry.KeepAlive()
		}
	}
	return nil

}

// FileServer registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func FileServer(pattern string, root string) {
	app.Router.Handler(http.MethodGet, pattern+"*filepath", NoLookDirHandler(http.StripPrefix(pattern, http.FileServer(http.Dir(root)))))
}

//NoLookDirHandler disable directory look
func NoLookDirHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		r.ParseForm()
		fn, ok := r.Form["rawName"]
		if ok && fn != nil && strings.TrimSpace(fn[0]) != "" {
			w.Header().Add("Content-Type", "application/octet-stream")
			w.Header().Add("Content-Disposition", "attachment;filename=\""+fn[0]+"\"")
		}
		h.ServeHTTP(w, r)
	})
}

//NotFoundHandler not found
type NotFoundHandler struct {
}

//ServeHTTP not found handler
func (NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	logs.Error("not-found",
		logs.String("url", r.RequestURI),
		logs.String("method", r.Method),
		logs.Int("status code", http.StatusNotFound),
		logs.String("status text", http.StatusText(http.StatusNotFound)),
	)
}

//MethodNotAllowedHandler method Not Allowed
type MethodNotAllowedHandler struct {
}

//ServeHTTP method Not Allowed handler
func (MethodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	logs.Error("method-not-allowed",
		logs.String("url", r.RequestURI),
		logs.String("method", r.Method),
		logs.Int("status code", http.StatusMethodNotAllowed),
		logs.String("status text", http.StatusText(http.StatusMethodNotAllowed)),
	)
}

//MethodOptionsHandler method Options
type MethodOptionsHandler struct {
}

//ServeHTTP method Options handler
func (MethodOptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allowOrigin := r.Header.Get("Origin")
	allowHeaders := r.Header.Get("Access-Control-Request-Headers")
	logs.Info("options",
		logs.String("url", r.RequestURI),
		logs.String("origin", allowOrigin),
		logs.String("host", r.Host),
		logs.String("referer", r.Referer()),
	)
	if app.cors.AllowHeaders != "" || allowHeaders == "" {
		w.Header().Set("Access-Control-Allow-Headers", app.cors.AllowHeaders)
		w.Header().Set("Access-Control-Expose-Headers", app.cors.AllowHeaders)
	} else {
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		w.Header().Set("Access-Control-Expose-Headers", allowHeaders)
	}
	if app.cors.AllowOrigin != "" || allowOrigin == "" {
		w.Header().Set("Access-Control-Allow-Origin", app.cors.AllowOrigin)
		w.Header().Set("Vary", app.cors.AllowOrigin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Vary", allowOrigin)
	}
	w.Header().Set("Access-Control-Allow-Methods", app.cors.AllowMethods)
	w.Header().Set("Access-Control-Max-Age", app.cors.MaxAge)
	w.Header().Set("Access-Control-Allow-Credentials", app.cors.AllowCredentials)
}

// doHandle registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func doHandle(pattern *Pattern) {
	if pattern.Fn == nil {
		return
	}
	//app.Router.HandlerFunc(method,pattern)
	app.Router.Handle(pattern.Method, pattern.Pattern, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logs.Info("access-start",
			logs.String("url", r.RequestURI),
			logs.String("method", r.Method),
		)
		st := time.Now()
		allowOrigin := r.Header.Get("Origin")
		if app.cors.AllowOrigin != "" || allowOrigin == "" {
			w.Header().Set("Access-Control-Allow-Origin", app.cors.AllowOrigin)
			w.Header().Set("Vary", app.cors.AllowOrigin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Vary", allowOrigin)
		}
		w.Header().Set("Access-Control-Allow-Credentials", app.cors.AllowCredentials)
		if pattern.Pattern == "/" && r.URL.Path != pattern.Pattern {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, http.StatusText(http.StatusNotFound))
			goto end
		}
		{
			ctx := app.pool.Get().(*Context)
			ctx.Reset()
			// defer app.pool.Put(ctx)
			defer func() {
				if ctx.Session != nil {
					//commit session data
					go ctx.Session.Commit()
				}
				app.pool.Put(ctx)
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
					panicCaller := zapcore.NewEntryCaller(runtime.Caller(4)).TrimmedPath()
					logs.ErrorWithCaller("access-panic",
						logs.String("caller", panicCaller),
						logs.Any("error", err),
						logs.String("url", r.RequestURI),
						logs.String("method", r.Method),
						logs.String("cost", time.Since(st).String()),
					)
				}
			}()

			ctx.Router = pattern
			ctx.In = r
			ctx.Accept = r.Header.Get("Accept")
			if ctx.Accept == "" || strings.HasPrefix(ctx.Accept, "application/json") {
				ctx.KindAccept = KindAcceptJSON
			} else if strings.HasPrefix(ctx.Accept, "application/xml") {
				ctx.KindAccept = KindAcceptXML
			} else if strings.HasPrefix(ctx.Accept, "application/x+yaml") {
				ctx.KindAccept = KindAcceptYAML
			}
			ctx.Out = w
			ctx.Params = ps
			ctx.NeedAuthorization = pattern.authorization

			s, err := session.Start(w, r)
			if err == nil && s != nil {
				ctx.Session = s
			}

			if pattern.authorization && app.Authorization != nil && app.Authorization(ctx) != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, http.StatusText(http.StatusUnauthorized))
				goto end
			}

			if ctx.NeedAuthorization {
				ctx.IsAuthorization = true
			}

			if app.Before != nil && app.Before(ctx) != nil {
				w.WriteHeader(http.StatusPreconditionFailed)
				fmt.Fprint(w, http.StatusText(http.StatusPreconditionFailed))
				goto end
			}

			pattern.Fn(ctx)

			if app.After != nil {
				app.After(ctx)
			}
		}
	end:
		logs.Info("access-end",
			logs.String("url", r.RequestURI),
			logs.String("method", r.Method),
			logs.String("cost", time.Since(st).String()),
		)
	})
}

//Serve use config(auto TLS) to start app
func Serve() bool {
	c := startServe()
	if c == nil {
		return false
	}
	if c.CertFile == "" {
		Run(c.Addr)
	} else {
		RunWithTLS(c.Addr, c.CertFile, c.KeyFile)
	}
	return true
}

//ServeWithAddr use addr to start app
func ServeWithAddr(addr string) bool {
	c := startServe()
	if c == nil {
		return false
	}
	Run(addr)
	return true
}

//ServeTLSWithAddr use addr and certFile, keyFile to start app
func ServeTLSWithAddr(addr, certFile, keyFile string) bool {
	startServe()
	RunWithTLS(addr, certFile, keyFile)
	return true
}

func startServe() *conf.AppConf {
	c := conf.Conf()
	if !app.isCallCommand && !Command() || c == nil {
		return nil
	}
	//seting validate lang
	if c.Lang == "" {
		c.Lang = "en"
	}
	valid.Lang = c.Lang
	Debug(c.Debug)
	return c
}

// TryServe try to check the configuration can be turned on
// 1: is control commandline
// 2: config is ok
func TryServe() bool {
	ok := Command() && conf.OK()
	if ok {
		cs := conf.Confs()
		for _, c := range cs {
			if c.Addr != "" {
				err := doTryRun(c.Addr)
				if err != nil {
					continue
				}
			}
		}
	}
	return ok
}

//IsRuning return app is runing
func IsRuning() bool {
	return app.runing
}

//Run use addr to start app
func Run(addr string) {
	if addr == "" {
		logs.Error("run addr is empty",
			logs.String("address", addr))
		logs.Sync()
		os.Exit(-1)
	}
	if !app.isCallCommand && !Command() {
		return
	}
	doRun(addr, "", "")
}

//RunWithTLS use addr and cert to start app
func RunWithTLS(addr, certFile, keyFile string) {
	if certFile == "" {
		logs.Error("run with TLS certFile is empty",
			logs.String("address", addr),
			logs.String("certFile", certFile),
			logs.String("keyFile", keyFile))
		logs.Sync()
		os.Exit(-1)
	}
	if !app.isCallCommand && !Command() {
		return
	}
	app.tls = true
	doRun(addr, certFile, keyFile)
}

//doRun real run app
func doRun(addr, certFile, keyFile string) {
	defer clear()
	Router()
	app.Addr = addr
	app.CertFile = certFile
	app.KeyFile = keyFile
	err := tryRun()
	if err == nil {
		logs.Info("run",
			logs.String("address", app.Addr))
		errMsg := ""
		if certFile == "" {
			err = app.ListenAndServe()
			errMsg = "listenAndServe"
		} else {
			err = app.ListenAndServeTLS()
			errMsg = "ListenAndServeTLS"
		}
		if err != nil {
			fmt.Println(errMsg + " error=" + err.Error())
			logs.Errors(errMsg, err)
			logs.Sync()
			os.Exit(0)
		}
		logs.Info("finish")
	} else {
		logs.Error("listen",
			logs.Err(err),
			logs.String("address", app.Addr),
			logs.String("certFile", app.CertFile),
			logs.String("keyFile", app.KeyFile))
		logs.Sync()
		os.Exit(-1)
	}
}

func tryRun() error {
	var err error
	for i := 0; i < 500; i++ {
		err = doTryRun(app.Addr)
		if err != nil {
			if i <= 10 {
				time.Sleep(2 * time.Microsecond)
			} else if i <= 50 {
				time.Sleep(20 * time.Microsecond)
			} else if i > 50 && i <= 100 {
				time.Sleep(1 * time.Millisecond)
			} else if i > 100 && i <= 250 {
				time.Sleep(2 * time.Millisecond)
			} else {
				time.Sleep(10 * time.Millisecond)
			}
			continue
		}
		break
	}
	return err
}

func doTryRun(add string) error {
	l, err := net.Listen("tcp", add)
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
		serviceInstall()
		r = false
	} else if flagStop {
		stop()
		r = false
	} else if flagReload {
		reload()
		r = false
	} else if flagDaemon {
		doDaemon()
	} else if isInstall {
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

//Debug set app debug status
func Debug(debug bool) {
	app.Debug = debug
}

func start() (bool, error) {
	if isMaster {
		doStart()
		checkWorkProcess()
		return false, nil
	}
	path := conf.Path()
	cmd := exec.Command(os.Args[0], "--master", "--start", "--conf="+path)
	cmd.Dir = AppDir()
	cmd.Start()
	// time.Sleep(200 * time.Millisecond)
	return false, nil
}

func serviceInstall() {
	if flagName == "" {
		flagName = AppName()
	}
	service, err := daemon.New(flagName, flagName+" --service")
	if err != nil {
		logs.Errors("service failed", err)
		return
	}
	status, err := service.Run(&daemonExecutable{})
	if err != nil {
		logs.Errors("service failed", err)
		return
	}
	logs.Info("service status", logs.String("status", status))
}

func doService() {
	doStart()
	go checkWorkProcess()
}

func doStart() error {
	appConfs := conf.Confs()
	if appConfs == nil || len(appConfs) <= 0 {
		logs.Error("not found conf")
		return errors.New("not found conf")
	}
	path := conf.Path()
	pid := strconv.Itoa(os.Getpid())
	app.pipeName = pid
	if flagService {
		logs.Info("service info", logs.String("path", path), logs.String("masterPid", pid))
	}
	app.cmd = []work{}
	for _, c := range appConfs {
		cmd := exec.Command(os.Args[0], "--daemon", "--appkey="+c.Key, "--pipe="+app.pipeName, "--conf="+path)
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		cmd.Dir = AppDir()
		err := cmd.Start()
		if err != nil {
			logs.Errors("create child process filed", err)
			return err
		}
		app.cmd = append(app.cmd, work{key: c.Key, cmd: cmd, runing: true})
	}
	if err := logPid(); err != nil {
		logs.Errors("logging error log pid", err)
		if !flagService {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func startWork(index int) *exec.Cmd {
	w := app.cmd[index]
	c := conf.Config(w.key)
	if c != nil {
		path := conf.Path()
		// pid := strconv.Itoa(os.Getpid())
		cmd := exec.Command(os.Args[0], "--daemon", "--appkey="+c.Key, "--pipe="+app.pipeName, "--conf="+path)
		// cmd.StdinPipe()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = AppDir()
		err := cmd.Start()
		if err != nil {
			return nil
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
		w := &app.cmd[i]
		if w.runing {
			go func(wc *work) {
				w.cmd.Wait()
				c <- struct{}{}
			}(w)
			l++
		}
	}
	for i := 0; i < lg; i++ {
		<-c
		w := &app.cmd[i]
		exitCode := ""
		if w.cmd.ProcessState != nil {
			exitCode = strconv.Itoa(w.cmd.ProcessState.ExitCode())
		}
		if flagService {
			logs.Error("has work process exited", logs.String("exitCode", exitCode))
		} else {
			fmt.Println("has work process exited,exit code=" + exitCode)
		}
		w.runing = false
		if !app.runing {
			break
		}
	}
	if flagService {
		logs.Error("exited check work process")
	} else {
		fmt.Println("exited check work process")
	}
	clear()
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
	start()
}

func serviceStop() {
	pids := getWorkPids()
	for _, pid := range pids {
		sendSignal(syscall.SIGINT, pid)
	}
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

func doDaemon() {
	app.Daemon = true
	go signalListen()
}

func install() {
	if flagName == "" {
		flagName = AppName()
	}
	var agrs = []string{"--service", "--force", "--conf=" + flagConf}

	service, err := daemon.New(flagName, flagName+" service")
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
	service, err := daemon.New(flagName, flagName+" service")
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
			logs.Info("signal", logs.String("signalText", s.String()))
			signal.Stop(c)
			err := Shutdown(nil)
			if err != nil {
				logs.Errors("shutdown error", err)
			} else {
				logs.Info("shutdown success")
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
	if isMaster {
		mPid := getMasterPidWithFile()
		if mPid == os.Getpid() {
			pidPath := os.Args[0] + ".pid"
			return os.Remove(pidPath)
		}
	}
	return nil
}

//getMasterPidWithFile return master pid form pid log
func getMasterPidWithFile() int {
	pidPath := os.Args[0] + ".pid"
	f, _ := os.Open(pidPath)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return 0
	}
	c := string(data)
	cs := strings.Split(c, "|")
	if len(cs) < 2 {
		return 0
	}
	c = cs[0]
	v, err := strconv.Atoi(c)
	if err != nil {
		return 0
	}
	return v
}

//getWorkPids work pids
func getWorkPids() []int {
	pids := getWorkPidsFormMemory()
	if pids == nil || len(pids) <= 0 {
		pids = getWorkPidsFormFile()
	}
	return pids
}

//getWorkPidsFormMemory get work pids form the memory
func getWorkPidsFormMemory() []int {
	pids := []int{}
	if app.cmd != nil {
		lg := len(app.cmd)
		for i := 0; i < lg; i++ {
			o := app.cmd[i]
			if o.cmd != nil && o.cmd.Process != nil && o.cmd.ProcessState != nil && o.cmd.ProcessState.Exited() && o.runing {
				pids = append(pids, o.cmd.Process.Pid)
			}
		}
	}
	return pids
}

//getWorkPidsFormFile get work pids form the file
func getWorkPidsFormFile() []int {
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

//True return a * bool
func True() *bool {
	tv := true
	return &tv
}

//False return a * bool
func False() *bool {
	fv := false
	return &fv
}

//String return a * string
func String(s string) *string {
	return &s
}

//ID create Unique ID
func ID() int64 {
	if app.id != nil {
		return app.id.GenerateWithInt64()
	}
	return 0
}

//GUID create GUID
func GUID() string {
	return guid.GUID()
}

/*conf*/

//RegistConf handler  conf
func RegistConf(handle conf.ConfingHandle, finish ...conf.FinishHandle) {
	conf.Register(handle, finish...)
}

//Conf returns the current app config
func Conf() *conf.AppConf {
	return conf.Conf()
}

//UserConf  returns the current user config
func UserConf() interface{} {
	return conf.UserConf()
}

//FileDir if app config configuration fileDir return it，orherwise return app exec path
func FileDir() string {
	return conf.FileDir()
}

//Trans translator
func Trans(language, key string, param ...string) string {
	return lang.Trans(language, key, param...)
}

//Transf translator
func Transf(language, key string, param ...interface{}) string {
	var ps []string
	if param != nil {
		ps = make([]string, 0, len(param))
		for _, v := range param {
			ps = append(ps, fmt.Sprint(v))
		}
	}
	return lang.Trans(language, key, ps...)
}

//LangFile translator file
func LangFile(file string) error {
	return lang.File(file)
}

//LangDir translator dir
func LangDir(dir string) error {
	return lang.Dir(dir)
}

//clear res
func clear() {
	if isClear {
		return
	}
	isClear = true
	logs.Clear()
	removePid()
	if app.registry != nil {
		app.registry.Stop()
	}
}

//parseCommandLine parse commandLine
func parseCommandLine() {
	doParseCommandLine()
	if flagName != "" {
		isInstall = true
	}
	if !isInstall {
		for _, k := range os.Args[1:] {
			if strings.HasPrefix(k, "--install") || strings.HasPrefix(k, "-i") {
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

}

func doParseCommandLine() {
	usageTemplate := `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}} 

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

	var cmd = &cobra.Command{
		Use:                   "app command args",
		Short:                 "",
		Long:                  "",
		SilenceErrors:         true,
		Example:               "app --start",
		DisableAutoGenTag:     true,
		DisableFlagsInUseLine: true,
		DisableSuggestions:    true,
		TraverseChildren:      true,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.Flags().BoolVarP(&flagDevelop, "develop", "d", flagDevelop, "run in develop(develop environment)")
	cmd.Flags().BoolVarP(&flagStart, "start", "s", flagStart, "run in background")
	cmd.Flags().BoolVarP(&flagStop, "stop", "e", flagStop, "graceful for stop")
	cmd.Flags().BoolVarP(&flagReload, "reload", "r", flagReload, "graceful for reload")
	cmd.Flags().BoolVarP(&isMigration, "migration", "m", isMigration, "migration or initial system")
	cmd.Flags().StringVarP(&flagConf, "conf", "c", flagConf, "config path(default is ./config.conf)")
	cmd.Flags().BoolVarP(&isInstall, "install", "i", isInstall, "install to service")
	cmd.Flags().BoolVarP(&isUninstall, "uninstall", "u", isUninstall, "uninstall for service")
	cmd.Flags().BoolVar(&flagDaemon, "daemon", flagDaemon, "")
	cmd.Flags().BoolVarP(&isForce, "force", "f", isForce, "")
	cmd.Flags().BoolVar(&flagService, "service", flagService, "")
	cmd.Flags().BoolVar(&isMaster, "master", isMaster, "")
	cmd.Flags().StringVarP(&flagName, "name", "n", flagName, "")
	cmd.Flags().StringVarP(&flagAppKey, "appkey", "k", flagAppKey, "app key")
	cmd.Flags().StringVarP(&flagPipe, "pipe", "p", flagPipe, "pipe name")
	cmd.Flags().IntVar(&flagPid, "pid", flagPid, "")
	cmd.SetUsageTemplate(usageTemplate)
	err := cmd.Execute()
	if err != nil {
		os.Exit(0)
	}
	conf.Command(&flagConf, &flagAppKey)

}
