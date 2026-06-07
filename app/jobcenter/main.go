// Package main 是 jobcenter 服务的入口点。
//
// jobcenter 是一个后台任务处理服务，主要职责包括：
//   - Kafka 消息消费：处理提现申请事件，执行链上转账
//   - 定时任务调度：同步汇率、K线数据等市场信息
//
// 该服务采用 go-zero 框架的 ServiceGroup 模式，将 Kafka Consumer 和
// 定时任务统一管理，共享依赖注入的 ServiceContext。
package main

import (
	"flag"
	"fmt"

	"mscoin_go/app/jobcenter/internal/config"
	"mscoin_go/app/jobcenter/internal/consumer"
	"mscoin_go/app/jobcenter/internal/svc"
	"mscoin_go/app/jobcenter/internal/task"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

// configFile 是命令行参数，指定配置文件路径。
// 默认值为 "etc/jobcenter.yaml"，可通过 -f 参数覆盖。
var configFile = flag.String("f", "etc/jobcenter.yaml", "配置文件路径")

// main 是 jobcenter 服务的启动入口。
//
// 启动流程：
//  1. 解析命令行参数，加载配置文件
//  2. 初始化 ServiceContext，建立数据库、缓存、RPC 客户端等依赖
//  3. 创建 Kafka Consumer（提现消费）和定时任务服务
//  4. 将所有服务加入 ServiceGroup 统一管理生命周期
//  5. 阻塞等待服务结束，优雅关闭资源
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.ServiceConf.MustSetUp()

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	withdrawConsumer, err := consumer.NewWithdrawConsumer(ctx)
	if err != nil {
		panic(err)
	}
	taskService := task.NewService(ctx)

	group := service.NewServiceGroup()
	group.Add(withdrawConsumer)
	group.Add(taskService)

	fmt.Printf("Starting jobcenter service %s...\n", c.Name)
	group.Start()
}
