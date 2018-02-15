package main

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func init() {
	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		lvl, e := log.ParseLevel(logLevel)
		if e != nil {
			log.Fatal("failed to set log level: ", e)
		}
		log.SetLevel(lvl)
	}
	if os.Getenv("LOG_FORMAT") == "json" {
		log.SetFormatter(
			&log.JSONFormatter{
				FieldMap: log.FieldMap{
					log.FieldKeyMsg:  "message",
					log.FieldKeyTime: "@timestamp",
				},
			})
	}
	log.SetOutput(os.Stdout)
}

func main() {
	tc := tasks()
	for t := range tc {
		if t.Status.State == "running" {
			s := service(t.ServiceID)
			// get ports from servic e labels, means we need to change orchestra to populate those
			log.Info("running task ", t.ID, " belongs to service ", s.Spec.Name, " labels ", s.Spec.Labels)
			// client, _ := consulapi.NewClient(consulapi.DefaultConfig())
			// catalog := client.Catalog()
			// consulRegistration := createConsulRegistration(s swarm.Service, t swarm.Task)
			// _, e := catalog.Register(consulRegistration, nil)
			// if e == nil {
			// 	log.Info("Service registered in Consul")
			// } else {
			// 	log.Error("Connection to consul failed: ", e)
			// }
		}
	}
}

// func createConsulRegistration() *consulapi.CatalogRegistration {
//bla
// }

func tasks() <-chan swarm.Task {
	c := make(chan swarm.Task)
	go func() {
		docker, _ := client.NewEnvClient()
		tl, e := docker.TaskList(context.Background(), types.TaskListOptions{})
		if e != nil {
			log.Error("failed to fetch data from docker. ", e)
		}
		for _, t := range tl {
			c <- t
		}
		close(c)
	}()
	return c
}

func service(id string) swarm.Service {
	var s swarm.Service
	docker, _ := client.NewEnvClient()
	l, e := docker.ServiceList(context.Background(), types.ServiceListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "id",
				Value: id,
			},
		),
	})
	if e != nil {
		log.Error("failed to get service list from docker. ", e)
	}
	for _, v := range l {
		s = v
	}
	return s
}

func node(id string) swarm.Node {
	var n swarm.Node
	docker, _ := client.NewEnvClient()
	l, e := docker.NodeList(context.Background(), types.NodeListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "id",
				Value: id,
			},
		),
	})
	if e != nil {
		log.Error("failed to fetch data from docker. ", e)
	}
	for _, v := range l {
		n = v
	}
	return n
}
