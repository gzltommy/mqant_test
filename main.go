package main

import (
	"fmt"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
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
	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"192.168.24.147:8500"}
	})

	nc, err := nats.Connect("nats://192.168.24.147:4222", nats.MaxReconnects(10000))
	if err != nil {
		log.Error("nats error %v", err)
		return
	}
	app := mqant.CreateApp(
		module.Debug(true),  //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
		module.Nats(nc),     //指定nats rpc
		module.Registry(rs), //指定服务发现
	)

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
