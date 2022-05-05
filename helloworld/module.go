package helloworld

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/server"
	"time"
)

var Module = func() module.Module {
	this := new(HellWorld)
	return this
}

type HellWorld struct {
	basemodule.BaseModule
}

func (self *HellWorld) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "helloworld"
}

//模块（服务）启动是,会自动注册模块 Version() 的返回值作为服务的版本
func (self *HellWorld) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *HellWorld) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings,
		server.RegisterInterval(15*time.Second), // 15 秒注册一次
		server.RegisterTTL(30*time.Second),      // 注册有效时间为 30 秒
		server.Id("mynode001"),                  // 手动指定一个节点id
	)

	// 设置一些元数据
	//元数据是节点级别的,且可以随时修改,利用好它可以灵活的实现定制化的服务发现 比如实现灰度发布,熔断策略等等
	self.GetServer().Options().Metadata["state"] = "alive"

	self.GetServer().RegisterGO("/say/hi", self.say) // 将 handler 注册到模块中
	self.GetServer().RegisterGO("HD_say", self.gatesay)
	log.Info("%v模块初始化完成...", self.GetType())
}

func (self *HellWorld) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", self.GetType())
	log.Info("%v say hello world...", self.GetType())
	<-closeSig
	log.Info("%v模块已停止...", self.GetType())
}

func (self *HellWorld) OnDestroy() {
	//一定别忘了继承
	//self.BaseModule.OnDestroy()
	log.Info("%v模块已回收...", self.GetType())
}

// 新增 handler 函数
func (self *HellWorld) say(name string) (r string, err error) {
	return fmt.Sprintf("hi %v", name), nil
}

func (self *HellWorld) gatesay(session gate.Session, msg map[string]interface{}) (r string, err error) {
	//主动给客户端发送消息
	session.Send("/gate/send/test", []byte(fmt.Sprintf("send hi to %v", msg["name"])))
	return fmt.Sprintf("hi %v 你在网关 %v", msg["name"], session.GetServerId()), nil
}
