package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/sirupsen/logrus"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kr/pretty"

	"github.com/fortinet-solutions-cse/fortiweb_go_client"
)

var fwb = &fortiwebclient.FortiWebClient{
	URL:      "https://192.168.122.40:90/",
	Username: "admin",
	Password: "",
}

func deleteAll() error {

	fmt.Println("Removing All K8S resources")

	deleteAllK8SContentRoutingPolicies()
	deleteAllK8SServerPoolRules()
	deleteAllK8SServerPolicies()
	deleteAllK8SVirtualServers()

	return nil

}

func deleteAllK8SContentRoutingPolicies() error {
	response, error := fwb.DoGet("api/v1.0/ServerObjects/Server/HTTPContentRoutingPolicy/K8S_HTTP_Content_Routing_Policy")

	if error != nil {
		return error
	}

	bodyByteArray, error := ioutil.ReadAll(response.Body)

	var bodyMap interface{}

	err := json.Unmarshal(bodyByteArray, &bodyMap)

	if err != nil {
		return err
	}

	contentRoutingPolicyList := reflect.ValueOf(bodyMap)

	for i := 0; i < contentRoutingPolicyList.Len(); i++ {

		value := contentRoutingPolicyList.Index(i).Interface()
		myMap := value.(map[string]interface{})

		fwb.DeleteContentRoutingPolicy(myMap["name"].(string))
	}

	return nil

}

func deleteAllK8SServerPoolRules() error {

	return nil
}

func deleteAllK8SServerPolicies() error {
	return nil
}

func deleteAllK8SVirtualServers() error {
	response, error := fwb.DoGet("api/v1.0/ServerObjects/Server/VirtualServer")

	if error != nil {
		return error
	}

	bodyByteArray, error := ioutil.ReadAll(response.Body)

	var bodyMap interface{}

	err := json.Unmarshal(bodyByteArray, &bodyMap)

	if err != nil {
		return err
	}

	virtualServerList := reflect.ValueOf(bodyMap)

	for i := 0; i < virtualServerList.Len(); i++ {

		value := virtualServerList.Index(i).Interface()
		myMap := value.(map[string]interface{})

		fwb.DeleteVirtualServer(myMap["name"].(string))
	}

	return nil

}

func transformIngressToFWB(ingress v1beta1.Ingress) {

	fmt.Println("Transforming ingress:" + ingress.GetName())

	//Delete All Resources created previously from K8S in FWB
	deleteAll()

	fmt.Println("Creating Virtual Server...")
	fwb.CreateVirtualServer("K8S_virtual_server",
		"", "", "port1",
		true, true)

	fmt.Println("Creating Server Policy...")
	fwb.CreateServerPolicy("K8S_Server_Policy",
		"K8S_virtual_server", "",
		"HTTP", "", "", "",
		fortiwebclient.HTTPContentRouting, 8192,
		false, false, false, false, false)

	for _, rule := range ingress.Spec.Rules {
		transformIngressRuleToFWB(rule)
	}
}

func transformIngressRuleToFWB(rule v1beta1.IngressRule) {
	transformHTTPRuleToFWB(rule.Host, rule.HTTP)
}

func transformHTTPRuleToFWB(host string, httpRule *v1beta1.HTTPIngressRuleValue) {

	for _, path := range httpRule.Paths {
		transformHTTPIngressPathToFWB(host, path)
	}
}

func getNodePortFromService(service string) int32 {

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config")

	clientset, err := getClient(kubeconfig)
	if err != nil {
		fmt.Println(strings.Join([]string{"Error:", err.Error()}, ""))
		logrus.Error(err)
		os.Exit(-1)
	}

	services, error := clientset.CoreV1().Services("default").List(v1.ListOptions{LabelSelector: "run=forum-webserver"})

	if error != nil || len(services.Items) == 0 {
		fmt.Println("Error getting services. Exiting...")
		os.Exit(-1)
	}

	fmt.Println("Returning:" + strconv.FormatInt(int64(services.Items[0].Spec.Ports[0].NodePort), 10))
	return services.Items[0].Spec.Ports[0].NodePort
}

func transformHTTPIngressPathToFWB(host string, ingressPath v1beta1.HTTPIngressPath) {

	path := ingressPath.Path
	service := ingressPath.Backend.ServiceName

	// Get port number of Service
	portNumber := getNodePortFromService(service)
	// Get list of nodes

	// CreateServerPool
	fmt.Println("Creating Server Pool...")
	fwb.CreateServerPool("K8S_Server_Pool_"+fwb.SafeURL(service),
		fortiwebclient.ServerBalance,
		fortiwebclient.ReverseProxy,
		fortiwebclient.RoundRobin,
		"")

	// CreateServerPoolRule
	fmt.Println("Creating Server Pool Rule 1...")
	fwb.CreateServerPoolRule("K8S_Server_Pool_"+fwb.SafeURL(service), "10.192.0.3", portNumber, 2, 0)
	fmt.Println("Creating Server Pool Rule 2...")
	fwb.CreateServerPoolRule("K8S_Server_Pool_"+fwb.SafeURL(service), "10.192.0.4", portNumber, 2, 0)

	// CreateHTTPContentRoutingPolicy
	fmt.Println("Creating HTTP Content Routing Policy...")
	fwb.CreateHTTPContentRoutingPolicy(fwb.SafeURL("K8S_HTTP_Content_Routing_Policy_"+host+"_"+path),
		fwb.SafeURL("K8S_Server_Pool_"+service),
		"(  )")

	// CreateHTTPContentRoutingUsingHost
	fmt.Println("Creating HTTP Content Routing for Host...")
	fwb.CreateHTTPContentRoutingUsingHost(fwb.SafeURL("K8S_HTTP_Content_Routing_Policy_"+host+"_"+path),
		host,
		3,
		fortiwebclient.AND)

	// CreateHTTPContentRoutingUsingURL
	fmt.Println("Creating HTTP Content Routing for URL...")
	fwb.CreateHTTPContentRoutingUsingURL(fwb.SafeURL("K8S_HTTP_Content_Routing_Policy_"+host+"_"+path),
		strings.Replace(path, "/", " ", -1),
		3,
		fortiwebclient.OR)

}

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

	fmt.Print("\n Getting Ingress: \n\n")

	ingresses, error := clientset.ExtensionsV1beta1().Ingresses("default").List(v1.ListOptions{})

	if error != nil {
		fmt.Println("Error getting ingress resources. Exiting...")
		os.Exit(-1)
	}

	for index, element := range ingresses.Items {

		fmt.Println(index)
		fmt.Println(pretty.Formatter(element))
	}

	fmt.Print("\nGetting Services: \n\n")

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

	fmt.Print("\nGetting Pods: \n\n")
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
	/*
		fwb := &fortiwebclient.FortiWebClient{
			URL:      "https://192.168.122.40:90/",
			Username: "admin",
			Password: "",
		}

		fmt.Println(fwb.GetStatus())

		fmt.Println("Creating Virtual Server...")
		fwb.CreateVirtualServer("K8S_virtual_server",
			"", "", "port1",
			true, true)

		fmt.Println("Creating Server Pool...")
		fwb.CreateServerPool("K8S_Server_Pool",
			fortiwebclient.ServerBalance,
			fortiwebclient.ReverseProxy,
			fortiwebclient.RoundRobin,
			"")

		fmt.Println("Creating Server Pool Rule 1...")
		fwb.CreateServerPoolRule("K8S_Server_Pool", "10.192.0.3", 30304, 2, 0)
		fmt.Println("Creating Server Pool Rule 2...")
		fwb.CreateServerPoolRule("K8S_Server_Pool", "10.192.0.4", 30304, 2, 0)

		fmt.Println("Creating HTTP Content Routing Policy...")
		fwb.CreateHTTPContentRoutingPolicy("K8S_HTTP_Content_Routing_Policy",
			"K8S_Server_Pool",
			"(  )")

		fmt.Println("Creating HTTP Content Routing for Host...")
		fwb.CreateHTTPContentRoutingUsingHost("K8S_HTTP_Content_Routing_Policy",
			"myhost",
			3,
			fortiwebclient.AND)

		fmt.Println("Creating HTTP Content Routing for URL...")
		fwb.CreateHTTPContentRoutingUsingURL("K8S_HTTP_Content_Routing_Policy",
			"myurl",
			3,
			fortiwebclient.OR)

		fmt.Println("Creating Server Policy...")
		fwb.CreateServerPolicy("K8S_Server_Policy",
			"K8S_virtual_server", "",
			"HTTP", "", "", "",
			fortiwebclient.HTTPContentRouting, 8192,
			false, false, false, false, false)

		fmt.Println("Creating Server Policy Content Rule...")
		fwb.CreateServerPolicyContentRule("K8S_Server_Policy",
			"K8S_Server_Policy_Content_Rule",
			"K8S_HTTP_Content_Routing_Policy",
			"",
			"",
			true,
			false)
	*/
	fmt.Print("\n Getting Ingress: \n\n")

	ingresses, error = clientset.ExtensionsV1beta1().Ingresses("default").List(v1.ListOptions{})

	if error != nil {
		fmt.Println("Error getting ingress resources. Exiting...")
		os.Exit(-1)
	}

	for _, ing := range ingresses.Items {
		transformIngressToFWB(ing)
	}

}
