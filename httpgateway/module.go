package httpgateway

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/httpgateway"
	go_api "github.com/liangdas/mqant/httpgateway/proto"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	mqrpc "github.com/liangdas/mqant/rpc"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	//" github.com/labstack/echo/v4"
)

var Module = func() module.Module {
	this := new(httpgate)
	return this
}

type httpgate struct {
	basemodule.BaseModule
}

func (self *httpgate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "httpgate"
}
func (self *httpgate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *httpgate) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)
	self.SetListener(self)

	//注册 handler(方案一)
	//self.GetServer().RegisterGO("/httpgate/topic", self.httpgateway)

	// 注册handler(方案二)
	/*
		利用 mqrpc.RPCListener 监听未实现的 handler，然后将请求通过 httptest 路由 web 框架中
		当 RPC 未找到已注册的 handler 时会调用 func NoFoundFunction(fn string)(*mqrpc.FunctionInfo,error)
	*/
	self.SetListener(self)
}

func (self *httpgate) NoFoundFunction(fn string) (*mqrpc.FunctionInfo, error) {
	return &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(self.httpgateway),
		Goroutine: true,
	}, nil
}
func (self *httpgate) BeforeHandle(fn string, callInfo *mqrpc.CallInfo) error {
	return nil
}
func (self *httpgate) OnTimeOut(fn string, Expired int64) {

}
func (self *httpgate) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {}
func (self *httpgate) OnComplete(fn string, callInfo *mqrpc.CallInfo, result *rpcpb.ResultInfo, exec_time int64) {
}

/*
	网关默认路由规则是从 URL.Path 的第一个段取出 moduleType
		/[moduleType]/path

	举例：
		http://127.0.0.1:8090/httpgate/topic
			moduleType httpgate
			hander /httpgate/topic
*/
func (self *httpgate) startHttpServer() *http.Server {
	srv := &http.Server{
		Addr: ":8090",
		// 方式一：使用默认路由规则
		//Handler: httpgateway.NewHandler(self.App), // 创建网关，使用默认路由规则

		// 方式二：编写自定义路由规则器
		Handler: httpgateway.NewHandler(self.App,
			httpgateway.SetRoute(func(app module.App, r *http.Request) (service *httpgateway.Service, e error) {
				return nil, nil
			})),
	}
	//http.Handle("/", httpgateway.NewHandler(self.App))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Info("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	// returning reference so caller can call Shutdown()
	return srv
}

func (self *httpgate) Run(closeSig chan bool) {
	log.Info("httpgate: starting HTTP server :8090")
	srv := self.startHttpServer()
	<-closeSig
	log.Info("httpgate: stopping HTTP server")
	// now close the server gracefully ("shutdown")
	// timeout could be given instead of nil as a https://golang.org/pkg/context/
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	log.Info("httpgate: done. exiting")
}

func (self *httpgate) OnDestroy() {
	//一定别忘了继承
	self.BaseModule.OnDestroy()
}

//网关转发 RPC 的 handler 定义为
//httpgateway 函数中，网络框架可以用 gin,echo,beego 等其他 web 框架替代
func (self *httpgate) httpgateway(request *go_api.Request) (*go_api.Response, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/httpgate/topic", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`hello world`))
	})

	req, err := http.NewRequest(request.Method, request.Url, strings.NewReader(request.Body))
	if err != nil {
		return nil, err
	}
	for _, v := range request.Header {
		req.Header.Set(v.Key, strings.Join(v.Values, ","))
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	resp := &go_api.Response{
		StatusCode: int32(rr.Code),
		Body:       rr.Body.String(),
		Header:     make(map[string]*go_api.Pair),
	}
	for key, vals := range rr.Header() {
		header, ok := resp.Header[key]
		if !ok {
			header = &go_api.Pair{
				Key: key,
			}
			resp.Header[key] = header
		}
		header.Values = vals
	}
	return resp, nil
}
