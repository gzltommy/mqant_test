package test_task

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangdas/armyant/task"
	"github.com/liangdas/armyant/work"
)

func NewWork(manager *Manager) *Work {
	this := new(Work)
	this.manager = manager
	//opts := this.GetDefaultOptions("tcp://127.0.0.1:3563")
	opts := this.GetDefaultOptions("ws://127.0.0.1:3653")
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Println("ConnectionLost", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		fmt.Println("OnConnectHandler")
	})
	err := this.Connect(opts)
	if err != nil {
		fmt.Println(err.Error())
	}

	this.On("/gate/send/test", func(client MQTT.Client, msg MQTT.Message) {
		fmt.Println(msg.Topic(), string(msg.Payload()))
	})
	return this
}

/**
Work 代表一个协程内具体执行任务工作者
*/
type Work struct {
	work.MqttWork
	manager *Manager
}

func (this *Work) UnmarshalResult(payload []byte) map[string]interface{} {
	rmsg := map[string]interface{}{}
	json.Unmarshal(payload, &rmsg)
	return rmsg["Result"].(map[string]interface{})
}

/**
每一次请求都会调用该函数,在该函数内实现具体请求操作

task:=task.Task{
        N:1000,    //一共请求次数，会被平均分配给每一个并发协程
        C:100,        //并发数
        //QPS:10,        //每一个并发平均每秒请求次数(限流) 不填代表不限流
}

N/C 可计算出每一个Work(协程) RunWorker将要调用的次数
*/
func (this *Work) RunWorker(t task.Task) {
	msg, err := this.Request("helloworld/HD_say", []byte(`{"name":"mqant"}`))
	if err != nil {
		return
	}

	fmt.Println(msg.Topic(), string(msg.Payload()))
}
func (this *Work) Init(t task.Task) {

}
func (this *Work) Close(t task.Task) {
	this.GetClient().Disconnect(0)
}
