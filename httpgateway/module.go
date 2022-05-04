package httpgateway

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	mqrpc "github.com/liangdas/mqant/rpc"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"net/http"
	"reflect"
)

var Module = func() module.Module {
	this := new(HttpGateWay)
	return this
}

type HttpGateWay struct {
	basemodule.BaseModule
}

func (self *HttpGateWay) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "HttpGateWay"
}
func (self *HttpGateWay) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *HttpGateWay) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)

	self.SetListener(self) // 设置监听器
}

func (self *HttpGateWay) startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":8081"}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Info("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	// returning reference so caller can call Shutdown()
	return srv
}

func (self *HttpGateWay) Run(closeSig chan bool) {
	log.Info("HttpGateWay: starting HTTP server :8080")
	srv := self.startHttpServer()
	<-closeSig
	log.Info("HttpGateWay: stopping HTTP server")
	// now close the server gracefully ("shutdown")
	// timeout could be given instead of nil as a https://golang.org/pkg/context/
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	log.Info("HttpGateWay: done. exiting")
}

func (self *HttpGateWay) OnDestroy() {
	//一定别忘了继承
	self.BaseModule.OnDestroy()
}

func (self *HttpGateWay) NoFoundFunction(fn string) (*mqrpc.FunctionInfo, error) {
	return &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(self.CloudFunction),
		Goroutine: true,
	}, nil
}

func (self *HttpGateWay) BeforeHandle(fn string, callInfo *mqrpc.CallInfo) error {
	return nil
}
func (self *HttpGateWay) OnTimeOut(fn string, Expired int64) {

}
func (self *HttpGateWay) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {}

/**
fn         方法名
params        参数
result        执行结果
exec_time     方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
*/
func (self *HttpGateWay) OnComplete(fn string, callInfo *mqrpc.CallInfo, result *rpcpb.ResultInfo, exec_time int64) {
}

func (self *HttpGateWay) CloudFunction(trace log.TraceSpan, request *http.Request) (*http.Response, error) {
	//e := echo.New()
	//ectest := httgatewaycontrollers.SetupRouter(self, e)
	//req, err := http.NewRequest(request.Method, request.Url, strings.NewReader(request.Body))
	//if err != nil {
	//	return nil, err
	//}
	//for _, v := range request.Header {
	//	req.Header.Set(v.Key, strings.Join(v.Values, ","))
	//}
	//rr := httptest.NewRecorder()
	//ectest.ServeHTTP(rr, req)
	//resp := &go_api.Response{
	//	StatusCode: int32(rr.Code),
	//	Body:       rr.Body.String(),
	//	Header:     make(map[string]*go_api.Pair),
	//}
	//for key, vals := range rr.Header() {
	//	header, ok := resp.Header[key]
	//	if !ok {
	//		header = &go_api.Pair{
	//			Key: key,
	//		}
	//		resp.Header[key] = header
	//	}
	//	header.Values = vals
	//}
	//return resp, nil
	return nil, nil
}
