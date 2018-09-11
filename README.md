# Kubernetes Ingress controller for FortiWeb

## Instructions 

### Install Docker in Docker

Use this repo: https://github.com/kubernetes-sigs/kubeadm-dind-cluster <br>
`wget https://cdn.rawgit.com/kubernetes-sigs/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.8.sh`<br>
`chmod +x dind-cluster-v1.8.sh`<br>
`./dind-cluster-v1.8.sh up`<br>


### External connectivity to port apiserver in port 8001 
`sudo sysctl -w net.ipv4.conf.ens160.route_localnet=1`<br>
`sudo iptables -t nat -I PREROUTING -p tcp --dport 8001 -j DNAT --to-destination 127.0.0.1:8080`


### Deploy a service and expose it
`kubectl run forum-webserver --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080 -r=1`<br>
`kubectl run image-webserver --image=docker.io/tutum/hello-world --port=80 -r=1`<br>

`kubectl expose deployment/forum-webserver --type="NodePort" --port 8080 --selector='run=forum-webserver'`<br>
`kubectl expose deployment/image-webserver --type="NodePort" --port 80 --selector='run=image-webserver'`<br>

`kubectl get services`<br>
`kubectl get pods`<br>
`kubectl get services/forum-webserver -o go-template='{{(index .spec.ports 0).nodePort}}'`<br>

### Deploy FortiWeb

Please use: https://github.com/fortinet-solutions-cse/testbeds

### Use fortinet-solutions-cse/testbeds/fortiweb
`wget https://raw.githubusercontent.com/fortinet-solutions-cse/testbeds/master/fortiweb/start_fwb_k8s.sh`<br>
`chmod +x ./start_fwb_k8s.sh`<br>
`./start_fwb_k8s.sh <fortiweb_kvm_qcow2_image>`<br>

### Change http port in FortiWeb to other than 80

Change the setting in System/Admin/Settings/HTTP 

### Create an ingress resource in K8S
`kubectl apply -f ingress_example_demo.yaml`<br>
`kubectl get ingress`<br>
`kubectl describe ingress`<br>

### Access pods
`kubectl get pods`<br>
`kubectl exec -it forum-webserver-d4f956cbc-v88lz bash`<br>

### Execute requests to fetch data from different pods

`wget --header "Host:foo.com" -O - http://192.168.122.40/path1`<br>
`wget --header "Host:foo.com" -O - http://192.168.122.40/path2`<br>
`wget --header "Host:bar.com" -O - http://192.168.122.40/path3`<br>
`wget --header "Host:bar.com" -O - http://192.168.122.40/path1`<br>
