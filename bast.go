//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/julienschmidt/httprouter"
)

var (
	usageline = `帮助:
	-h | -help                  显示帮助
	-develop                    以后开发模式启动(开发过程中配置)
	-start                      以后台启动(可以与conf同时使用)
	-stop                       平滑停止
	-reload                     平滑升级程序(可以与conf同时使用)
	-conf=your path/config.conf  配置文件路径 
	`
	flagDevelop, flagStart, flagStop, flagReload, flagDaemon bool
	flagConf                                                 string
	app                                                      *App
)

//App is application major data
type App struct {
	pool   sync.Pool
	Router *httprouter.Router
	Addr   string
	Server *http.Server
	Before BeforeHandle
	After  AfterHandle
	Debug  bool
	Daemon bool
}

//init application
func init() {
	parseCommandLine()
	app = &App{Server: &http.Server{}, Router: httprouter.New()}
	doHandle("OPTIONS", "/*filepath", nil)
	app.pool.New = func() interface{} {
		return &Context{}
	}
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
	f.BoolVar(&flagStart, "start", true, "")
	f.BoolVar(&flagStop, "stop", false, "")
	f.BoolVar(&flagReload, "reload", false, "")
	f.BoolVar(&flagDaemon, "daemon", false, "")
	f.StringVar(&flagConf, "conf", "", "")
	f.Parse(os.Args[1:])
	if flagDevelop || flagDaemon {
		flagStart = false
	}
	parseConf(f)
}

//BeforeHandle is before then request handler
type BeforeHandle func(ctx *Context) error

//AfterHandle is after then request handler
type AfterHandle func(ctx *Context) error

//Before 请求前处理程序
func Before(f BeforeHandle) {
	app.Before = f
}

//After 请求前处理程序
func After(f AfterHandle) {
	app.After = f
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
func Post(pattern string, f func(ctx *Context)) {
	doHandle("POST", pattern, f)
}

// Get registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Get(pattern string, f func(ctx *Context)) {
	doHandle("GET", pattern, f)
}

// FileServer registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func FileServer(pattern string, root string) {
	app.Router.Handler("GET", pattern+"*filepath", NoLookDirHandler(http.StripPrefix(pattern, http.FileServer(http.Dir(root)))))
}

//NoLookDirHandler 不启用目录浏览
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
func doHandle(method, pattern string, f func(ctx *Context)) {
	//app.Router.HandlerFunc(method,pattern)
	app.Router.Handle(method, pattern, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logs.Info(r.Method + ":" + r.RequestURI + "->start")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			//有使用自定义头 需要这个,Action, Module是例子
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization,Access-Control-Allow-Origin,Content-Length,Content-Type,BaseUrl")
			//
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
				ctx.ResponseWriter = w
				ctx.Request = r
				ctx.Params = ps
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
			logs.Info(r.Method + ":" + r.RequestURI + "->end=notAllowed")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, http.StatusText(http.StatusMethodNotAllowed))
		}
	})
}

//Run app
func Run(addr string) {
	if !Command() {
		return
	}
	defer clear()
	doRun(addr)
}

//Debug set app debug status
func Debug(debug bool) {
	app.Debug = debug
}

//doRun real run app
func doRun(addr string) {
	app.Addr = addr
	logs.Info("app-runing-addr=" + app.Addr)
	fmt.Println("app start")
	err := app.ListenAndServe()
	if err != nil {
		fmt.Println("app start error=" + err.Error())
		logs.Info("app-listenAndServe-error=" + err.Error()) //+ ",pid=" + strconv.Itoa(os.Getpid())
	}
	fmt.Println("app finish")
	logs.Info("app-finish")
}

//Command 解析命令行参数
func Command() bool {
	// args := os.Args[1:]
	// lg := len(args)
	r := true
	var err error
	if flagStart {
		start()
		err = errors.New("start child process")
		r = false
	} else if flagStop {
		stop()
		err = errors.New("stop child process")
		r = false
	} else if flagReload {
		reload()
		err = errors.New("inside child process for reload")
		r = false
	} else if flagDaemon {
		daemon()
	}
	if err != nil {
		fmt.Printf(err.Error())
	}
	return r
}

func start() {
	path := Conf()
	fmt.Println("start=" + path)
	cmd := exec.Command(os.Args[0], "-daemon", "-conf="+path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	if err := logPid(cmd.Process.Pid); err != nil {
		fmt.Printf(err.Error())
		cmd.Process.Kill()
	}
	if err := logMgr(); err != nil {
		fmt.Printf(err.Error())
		cmd.Process.Kill()
	}
	os.Exit(0)
}

func reload() {
	path := mgrPath()
	fmt.Println("reload=" + path)
	sendSignal(syscall.SIGINT, path)
	time.Sleep(30 * time.Millisecond)
	flagConf = path
	start()
}

func stop() {
	path := mgrPath()
	fmt.Println("stop=" + path)
	sendSignal(syscall.SIGINT, path)
	// fmt.Println("stop=" + path)
	os.Exit(0)
}

func mgrPath() string {
	path := ""
	pos := 0
	mls := mgrList()
	if mls != nil {
		lg := len(mls)
		if lg > 0 {
			if lg > 1 {
				fmt.Print("位置列表：\n")
				for i := 0; i < lg; i++ {
					fmt.Printf("    %d：", i+1)
					fmt.Println(mls[i])
				}
				fmt.Print("请输入位置序号：")
				for {
					_, err := fmt.Scanf("%d", &pos)
					if err != nil || pos <= 0 || pos > lg {
						fmt.Print("请输入正确位置序号：")
						continue
					}
					pos--
					break
				}
				path = mls[pos]
				mls = append(mls[:pos], mls[pos+1:]...)
			} else {
				path = mls[0]
				mls = nil
			}
		}
	}
	syncMgr(mls)
	return path
}

//Conf config path
func Conf() string {
	return flagConf
}

//AppDir app path
func AppDir() string {
	exPath := filepath.Dir(Conf())
	return exPath
}

//parseConf parse config path
func parseConf(f *flag.FlagSet) string {
	exPath := filepath.Dir(os.Args[0])
	fs := f.Lookup("conf")
	if fs != nil {
		flagConf = fs.Value.String()
	}
	// flag.StringVar(&conf, "conf", "", "")
	if flagConf == "" {
		cf := exPath + "/config.conf"
		flagConf = cf
	}
	return flagConf
}

func sendSignal(sig os.Signal, path ...string) error {
	pid := getPid(path...)
	pro, err := os.FindProcess(pid) //通过pid获取子进程
	if err != nil {
		return err
	}
	err = pro.Signal(sig) //给子进程发送信号使之结束
	if err != nil {
		return err
	}
	return nil
}

func daemon() {
	app.Daemon = true
	go signalListen()
}

func signalListen() {
	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c)
	for {
		s := <-c
		logs.Info("signal=" + s.String())
		if s == syscall.SIGINT {
			signal.Stop(c)
			logs.Info("signal=syscall.SIGINT")
			err := Shutdown(nil)
			if err != nil {
				logs.Info("shutdown-error=" + err.Error())
			} else {
				logs.Info("shutdown-ok=")
			}
			break
		}
	}
}

func logPid(pid int) error {
	pidPath := pidPath()
	f, err := os.OpenFile(pidPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return errors.New("创建pid文件失败")
	}
	defer f.Close()
	if _, err := f.Write([]byte(strconv.Itoa(pid))); err != nil {
		return errors.New("写如pid文件失败")
	}
	f.Sync()
	return nil
}

func logMgr() error {
	cf := Conf()
	mgrPath := os.Args[0] + ".mgr"
	mls := mgrList()
	if mls != nil {
		for _, v := range mls {
			if v == cf {
				return nil
			}
		}
	}
	f, err := os.OpenFile(mgrPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.New("创建mgrFile文件失败")
	}
	defer f.Close()
	if _, err := f.Write([]byte(cf + "\n")); err != nil {
		return errors.New("写如mgrFile文件失败")
	}
	f.Sync()
	return nil
}

func syncMgr(appList []string) error {
	mgrPath := os.Args[0] + ".mgr"
	cf := ""
	if appList != nil {
		for _, v := range appList {
			cf += v + "\n"
		}
	}
	// fmt.Println(mgrPath + "=" + cf)
	f, err := os.OpenFile(mgrPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return errors.New("同步mgrFile文件失败")
	}
	defer f.Close()
	if _, err := f.WriteString(cf); err != nil {
		return errors.New("同步mgrFile文件失败")
	}
	f.Sync()
	return nil
}

func mgrList() []string {
	mgrPath := os.Args[0] + ".mgr"
	f, err := os.OpenFile(mgrPath, os.O_RDONLY, 0666)
	if err != nil {
		return nil
	}
	buf := bufio.NewReader(f)
	ls := []string{}
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			ls = append(ls, line)
		}
		if err != nil {
			if err == io.EOF {
				return ls
			}
		}
	}
}

func removePid() error {
	pidPath := pidPath()
	return os.Remove(pidPath)
}

//getPid pid
func getPid(path ...string) int {
	ppath := pidPath(path...)
	f, _ := os.Open(ppath)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return 0
	}
	p := string(data)
	id, err := strconv.Atoi(p)
	if err != nil {
		return 0
	}
	return id
}

//pidPath pid filename path
func pidPath(path ...string) string {
	pidPath := ""
	if path != nil {
		pidPath = path[0]
	}
	if pidPath == "" {
		pidPath = Conf() + ".pid"
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
	logs.ClearLogger()
	ids.IDClear()
	removePid()
}
