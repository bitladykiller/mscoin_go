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

var configFile = flag.String("f", "etc/jobcenter.yaml", "the config file")

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
