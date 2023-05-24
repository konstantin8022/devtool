package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
)

func MustConnectToK8s() *kubernetes.Clientset {
	glog.Info("getting connection to k8s API")
	var config *rest.Config
	var err error
	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Exitf("failed to get k8s config: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Exitf("failed to connect to k8s API server: %v", err)
	}

	glog.Info("successfully connected to k8s API server!")
	return client
}
