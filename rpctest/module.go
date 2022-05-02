package rpctest

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
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
func (self *rpctest) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *rpctest) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)
	self.GetServer().RegisterGO("/test/proto", self.testProto)
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
