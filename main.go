package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

var (
	e error
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
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	catalog := client.Catalog()
	for {
		tc := tasks()
		for t := range tc {
			if t.Status.State == "running" {
				s := service(t.ServiceID)
				for _, v := range t.NetworksAttachments {
					if v.Network.Spec.Name != "ingress" {
						log.Debug("v is: ", v.Addresses, v.Network.Spec.Name)
						log.Info("running task ", t.ID, " belongs to service ", s.Spec.Name, " with labels ", s.Spec.Labels)
						m := convertPort(s.Spec.Labels)
						ip := strings.Split(v.Addresses[0], "/")[0]
						//
						for l, p := range m {
							ccr := createCatalogRegistration(s.Spec.Name, l, t.ID, ip, p)
							registerInConsul(ccr, catalog)
						}
					}
				}
			}
		}
		time.Sleep(time.Minute * 1)
	}
}

func registerInConsul(cr consulapi.CatalogRegistration, c *consulapi.Catalog) {
	_, e := c.Register(&cr, nil)
	if e != nil {
		log.Error("failed to register ", cr.Service.Service, " in consul: ", e)
	}
}

func createCatalogRegistration(name, label, taskID, ip string, p int) consulapi.CatalogRegistration {
	var (
		cas consulapi.AgentService
		ccr consulapi.CatalogRegistration
	)
	sp := strconv.Itoa(p)
	// cas.ID = s.Spec.Name + ":" + string(p)
	cas.Service = name + ":" + sp
	cas.Tags = []string{taskID, label, sp, "prometheus"}
	cas.Port = p
	cas.Address = ip

	// ccr.ID = s.Spec.Name + ":" + string(p)
	ccr.Node = "monitoring"
	ccr.Address = "127.0.0.1"
	ccr.TaggedAddresses = map[string]string{"wan": ip, "lan": ip}
	ccr.Datacenter = "dc1"
	ccr.Service = &cas
	log.Debug(ccr)
	return ccr
}

func convertPort(l map[string]string) map[string]int {
	m := make(map[string]int)
	for i, v := range l {
		log.Debug("debug: ", i, v)
		if strings.HasPrefix(i, "prometheus.metrics") {
			p, e := strconv.Atoi(v)
			if e != nil {
				log.Error("failed to convert label value for: ", i, v)
			}
			m[i] = p
		}
	}
	return m
}

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
