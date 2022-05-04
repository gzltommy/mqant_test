package rpctest

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/server"
	"time"
)

var Module = func() module.Module {
	this := new(rpctest)
	return this
}

type rpctest struct {
	basemodule.BaseModule
}

func (self *rpctest) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "rpctest"
}

//模块（服务）启动是,会自动注册模块 Version() 的返回值作为服务的版本
func (self *rpctest) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *rpctest) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings,
		server.RegisterInterval(15*time.Second), // 15 秒注册一次
		server.RegisterTTL(30*time.Second),      // 注册有效时间为 30 秒
		server.Id("mynode001"),                  // 手动指定一个节点id
	)

	// 设置一些元数据
	//元数据是节点级别的,且可以随时修改,利用好它可以灵活的实现定制化的服务发现 比如实现灰度发布,熔断策略等等
	self.GetServer().Options().Metadata["state"] = "alive"

	// 在模块中获取应用级别的自定义配置
	MongodbUrl := app.GetSettings().Settings["MongodbURL"].(string)
	MongodbDB := app.GetSettings().Settings["MongodbDB"].(string)
	_ = MongodbUrl
	_ = MongodbDB

	// 在模块中获取模块的自定义配置
	StaticPath := self.GetModuleSettings().Settings["StaticPath"].(string)
	Port := int(self.GetModuleSettings().Settings["Port"].(float64))
	_ = StaticPath
	_ = Port

	// 为该模块注册 handler
	// Register（单线程）/RegisterGO（多线程）
	self.GetServer().RegisterGO("/test/proto", self.testProto)
	self.GetServer().RegisterGO("/test/marshal", self.testMarshal)
}

func (self *rpctest) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", self.GetType())
	<-closeSig
}

func (self *rpctest) OnDestroy() {
	//一定别忘了继承
	self.BaseModule.OnDestroy()
}
func (self *rpctest) testProto(req *rpcpb.ResultInfo) (*rpcpb.ResultInfo, error) {
	r := &rpcpb.ResultInfo{Error: *proto.String(fmt.Sprintf("你说: %v", req.Error))}
	return r, nil
}

func (self *rpctest) testMarshal(req Req) (*Rsp, error) {
	r := &Rsp{Msg: fmt.Sprintf("你的ID：%v", req.Id)}
	return r, nil
}
