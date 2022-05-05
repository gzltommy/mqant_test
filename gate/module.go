package mgate

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/module"
)

var Module = func() module.Module {
	gate := new(Gate)
	return gate
}

type Gate struct {
	basegate.Gate //继承
}

func (this *Gate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Gate"
}
func (this *Gate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

func (this *Gate) OnInit(app module.App, settings *conf.ModuleSettings) {
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	this.Gate.OnInit(this, app, settings,
		gate.WsAddr(":3653"),
		gate.TcpAddr(":3563"),
	)
}
