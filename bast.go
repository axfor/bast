//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/julienschmidt/httprouter"
)

var (
	app *App
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
}

//init application
func init() {
	app = &App{Server: &http.Server{}, Router: httprouter.New()}
	//app.Router.HandleOPTIONS = true
	doHandle("OPTIONS", "/*filepath", nil)
	app.pool.New = func() interface{} {
		return &Context{}
	}
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
	return http.ListenAndServe(app.Addr, app.Router)
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
	err := app.ListenAndServe()
	if err != nil {
		//app.Server.Shutdown(nil)
		logs.Info("app-listenAndServe-error=" + err.Error())
	}
	logs.Info("app-finish")
}

//Shutdown app
func Shutdown(ctx context.Context) {
	app.Server.Shutdown(ctx)
}

//clear res
func clear() {
	logs.ClearLogger()
	ids.IDClear()
}
