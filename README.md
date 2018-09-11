#Kubernetes Ingress controller for FortiWeb

## Instructions 

### Install Docker in Docker

Use this repo: https://github.com/kubernetes-sigs/kubeadm-dind-cluster
`wget https://cdn.rawgit.com/kubernetes-sigs/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.8.sh`
`chmod +x dind-cluster-v1.8.sh`
`./dind-cluster-v1.8.sh up`


### External connectivity to port apiserver in port 8001 
`sudo sysctl -w net.ipv4.conf.ens160.route_localnet=1`
`sudo iptables -t nat -I PREROUTING -p tcp --dport 8001 -j DNAT --to-destination 127.0.0.1:8080`


### Deploy a service and expose it
`kubectl run forum-webserver --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080 -r=1`
`kubectl run image-webserver --image=docker.io/tutum/hello-world --port=80 -r=1`

`kubectl expose deployment/forum-webserver --type="NodePort" --port 8080 --selector='run=forum-webserver'`
`kubectl expose deployment/image-webserver --type="NodePort" --port 80 --selector='run=image-webserver'`

`kubectl get services`
`kubectl get pods`
`kubectl get services/forum-webserver -o go-template='{{(index .spec.ports 0).nodePort}}'`

### Deploy FortiWeb

Please use: https://github.com/fortinet-solutions-cse/testbeds

### Use fortinet-solutions-cse/testbeds/fortiweb
`wget https://raw.githubusercontent.com/fortinet-solutions-cse/testbeds/master/fortiweb/start_fwb_k8s.sh`
`chmod +x ./start_fwb_k8s.sh`
`./start_fwb_k8s.sh <fortiweb_kvm_qcow2_image>`

### Change http port in FortiWeb to other than 80

Change the setting in System/Admin/Settings/HTTP 

### Create an ingress resource in K8S
`kubectl apply -f ingress_example_demo.yaml`
`kubectl get ingress`
`kubectl describe ingress`

### Access pods
`kubectl get pods`
`kubectl exec -it forum-webserver-d4f956cbc-v88lz bash`

### Execute requests to fetch data from different pods

`wget --header "Host:foo.com" -O - http://192.168.122.40/path1`
`wget --header "Host:foo.com" -O - http://192.168.122.40/path2`
`wget --header "Host:bar.com" -O - http://192.168.122.40/path3`
`wget --header "Host:bar.com" -O - http://192.168.122.40/path1`
