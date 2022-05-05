package main

import (
	"flag"
	"fmt"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	mqrpc "github.com/liangdas/mqant/rpc"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
	"math/rand"
	"mqant_test/helloworld"
	"mqant_test/rpctest"
	"mqant_test/web"
	"strconv"
	"time"
)

type WeightNode struct {
	Node   *registry.Node
	Weight int
}

func main() {
	flag.Parse() //解析输入的参数
	/*
		mqant 会在工作路径下初始化未指定的设置：
		配置文件 {workdir}/bin/conf/server.json
		日志文件目录 {workdir}/bin/logs
		BI 日志文件目录 {workdir}/bin/bi
	*/
	wdPath := *flag.String("wd", "", "Server work directory")              // 进程的工作路径
	confPath := *flag.String("conf", "", "Server configuration file path") //指定配置文件
	processID := *flag.String("pid", "development", "Server ProcessID?")   //指定模块分组ID
	//Logdir = *flag.String("log", "", "Log file directory?")
	//BIdir = *flag.String("bi", "", "bi file directory?")

	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"192.168.24.147:8500"}
	})

	nc, err := nats.Connect("nats://192.168.24.147:4222", nats.MaxReconnects(10000))
	if err != nil {
		log.Error("nats error %v", err)
		return
	}
	/*
	   服务发现配置只能通过创建 app 代码设置,包涵:
	   		* nats 配置
	   		* 注册中心配置(consul,etcd)
	   		* 服务发现 TTL 和注册间隔
	*/

	app := mqant.CreateApp(
		module.Debug(true),  //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
		module.Nats(nc),     //指定nats rpc
		module.Registry(rs), //指定服务发现
		module.RegisterTTL(20*time.Second),
		module.RegisterInterval(10*time.Second),

		module.WorkDir(wdPath),
		module.Configure(confPath),
		module.ProcessID(processID),

		//我们通常希望能监控handler的具体执行情况,例如做监控报警等等
		//应用级别handler监控

		//调用方监控
		module.SetClientRPChandler(func(app module.App, server registry.Node, rpcinfo *rpcpb.RPCInfo, result interface{}, err string, exec_time int64) {

		}),

		//服务方监控
		module.SetServerRPCHandler(func(app module.App, module module.Module, callInfo *mqrpc.CallInfo) {

		}),
	)

	// 在应用中获取应用级别的自定义配置
	_ = app.OnConfigurationLoaded(func(app module.App) {
		MongodbUrl := app.GetSettings().Settings["MongodbURL"].(string)
		MongodbDB := app.GetSettings().Settings["MongodbDB"].(string)
		_ = MongodbUrl
		_ = MongodbDB
	})

	// 应用级别的节点选择策略
	_ = app.Options().Selector.Init(selector.SetStrategy(func(services []*registry.Service) selector.Next {
		var nodes []WeightNode
		// Filter the nodes for datacenter
		for _, service := range services {
			for _, node := range service.Nodes {
				weight := 100
				if w, ok := node.Metadata["weight"]; ok {
					wint, err := strconv.Atoi(w)
					if err == nil {
						weight = wint
					}
				}
				if state, ok := node.Metadata["state"]; ok {
					if state != "forbidden" {
						nodes = append(nodes, WeightNode{
							Node:   node,
							Weight: weight,
						})
					}
				} else {
					nodes = append(nodes, WeightNode{
						Node:   node,
						Weight: weight,
					})
				}
			}
		}
		//log.Info("services[0] $v",services[0].Nodes[0])
		return func() (*registry.Node, error) {
			if len(nodes) == 0 {
				return nil, fmt.Errorf("no node")
			}
			rand.Seed(time.Now().UnixNano())
			//按权重选
			total := 0
			for _, n := range nodes {
				total += n.Weight
			}
			if total > 0 {
				weight := rand.Intn(total)
				togo := 0
				for _, a := range nodes {
					if (togo <= weight) && (weight < (togo + a.Weight)) {
						return a.Node, nil
					} else {
						togo += a.Weight
					}
				}
			}
			//降级为随机
			index := rand.Intn(int(len(nodes)))
			return nodes[index].Node, nil
		}
	}))

	err = app.Run( //模块都需要加到入口列表中传入框架
		helloworld.Module(),
		web.Module(),
		rpctest.Module(),
	)
	if err != nil {
		log.Error(err.Error())
	}
}
