// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const waitForTelemetryInSeconds = 60

// Version is populated by make during build.
var version string

func initLogging() error {
	log.SetName("azure-npm")
	log.SetLevel(log.LevelInfo)
	if runtime.GOOS == "windows" {
		log.SetLogDirectory("/var/log/")
	}
	if err := log.SetTarget(log.TargetLogfile); err != nil {
		log.Logf("Failed to configure logging, err:%v.", err)
		return err
	}

	return nil
}

func main() {
	var err error

	defer func() {
		if r := recover(); r != nil {
			log.Logf("recovered from error: %v", err)
		}
	}()

	if err = initLogging(); err != nil {
		panic(err.Error())
	}

	var config *rest.Config
	if runtime.GOOS == "windows" {
		// Creates the out-of-cluster config
		kubeconfig := flag.String("kubeconfig", "C:\\k\\config", "(optional) absolute path to the kubeconfig file")
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		// Creates the in-cluster config
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Logf("clientset creation failed with error %v.", err)
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)

	npMgr := npm.NewNetworkPolicyManager(clientset, factory, version)

	//go npMgr.SendNpmTelemetry()

	//time.Sleep(time.Second * waitForTelemetryInSeconds)

	err = npMgr.Start(wait.NeverStop)
	if err != nil {
		log.Logf("npm failed with error %v.", err)
		panic(err.Error)
	}

	select {}
}
