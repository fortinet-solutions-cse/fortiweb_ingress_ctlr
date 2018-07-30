package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kr/pretty"

	"github.com/fortinet-solutions-cse/fortiweb_go_client"
)

func getClient(pathToCfg string) (*kubernetes.Clientset, error) {

	var config *rest.Config
	var err error
	if pathToCfg == "" {
		logrus.Info("Using in cluster config")
		config, err = rest.InClusterConfig()

	} else {
		logrus.Info("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

var clientset *kubernetes.Clientset
var controller cache.Controller
var store cache.Store

func _init() {

	fmt.Println("_init")
}

func init() {
	fmt.Println("init")

}
func main() {

	fmt.Println("main")

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "Kube config path for outside of cluster access",
		},
	}

	fmt.Println("Hi")
	logrus.Info("Welcome")

	var tag string

	tag = "Test"

	fmt.Println(tag)

	location := "/home/magonzalez/.kube/config"

	fmt.Println(location)

	clientset, err := getClient(location)
	if err != nil {
		fmt.Println(strings.Join([]string{"Error:", err.Error()}, ""))
		logrus.Error(err)
		os.Exit(-1)
	}

	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})

	if err != nil {
		fmt.Println("Error getting nodes")
		os.Exit(-1)
	}

	fmt.Println("Length:" + strconv.Itoa(len(nodes.Items)))

	if len(nodes.Items) > 0 {

		fmt.Println(nodes.Items)

	}

	fmt.Println("Beautified:")

	for index, element := range nodes.Items {

		fmt.Println("Id: " + strconv.Itoa(index))
		fmt.Println(element.Status.NodeInfo.MachineID)
		fmt.Println(element.Status.NodeInfo.KernelVersion)

	}

	fmt.Println("\n Getting Ingress: ")

	ingressses, error := clientset.ExtensionsV1beta1().Ingresses("default").List(v1.ListOptions{})

	if error != nil {
		fmt.Println("Error getting ingress resources. Exiting...")
		os.Exit(-1)
	}

	for index, element := range ingressses.Items {

		fmt.Println(index)
		fmt.Println(pretty.Formatter(element))
	}

	fmt.Println("\nGetting Services: \n")

	services, error := clientset.CoreV1().Services("default").List(v1.ListOptions{})

	if error != nil {
		fmt.Println("Error getting services. Exiting...")
		os.Exit(-1)
	}

	for index, element := range services.Items {

		fmt.Println(index)
		fmt.Println(pretty.Formatter(element.Name))
		fmt.Println(pretty.Formatter(element.Spec.Selector))

		selectors := element.Spec.Selector

		for key, value := range selectors {
			fmt.Println(key, " = ", value)
		}
	}

	fmt.Println("\nGetting Pods: \n")
	pods, error := clientset.CoreV1().Pods("default").List(v1.ListOptions{})

	if error != nil {
		fmt.Println("Error getting pods. Exiting...", error.Error())
	}

	for index, element := range pods.Items {

		fmt.Println(index, element)
		labels := element.Labels

		fmt.Println(labels)

		fmt.Println("Pod name: ", element.Name)
		fmt.Println("Node IP: ", element.Status.HostIP)

	}

	fwb := &fortiwebclient.FortiWebClient{
		URL:      "https://192.168.122.40:90/",
		Username: "admin",
		Password: "",
	}

	fmt.Println(fwb.GetStatus())

	body := map[string]interface{}{
		"name":           "K8S_virtual_server5",
		"ipv4Address":    "0.0.0.0/0.0.0.0",
		"ipv6Address":    "::/0",
		"interface":      "port1",
		"useInterfaceIP": true,
		"enable":         true,
		"can_delete":     true,
	}

	fmt.Println(body)

	jsonByte, err := json.Marshal(body)
	fwb.CreateVirtualServer(string(jsonByte))

}
