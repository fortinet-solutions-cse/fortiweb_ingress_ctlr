#Install Docker

#Install kubectl

#Install Docker in Docker:
# https://github.com/kubernetes-sigs/kubeadm-dind-cluster
wget https://cdn.rawgit.com/kubernetes-sigs/kubeadm-dind-cluster/master/fixed/dind-cluster-v1.8.sh
chmod +x dind-cluster-v1.8.sh
./dind-cluster-v1.8.sh up


# External connectivity to port apiserver in port 8001 
sudo sysctl -w net.ipv4.conf.ens160.route_localnet=1
sudo iptables -t nat -I PREROUTING -p tcp --dport 8001 -j DNAT --to-destination 127.0.0.1:8080


# Deploy a service and expose it
kubectl run forum-webserver --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080 -r=1
kubectl run image-webserver --image=docker.io/tutum/hello-world --port=80 -r=1

kubectl expose deployment/forum-webserver --type="NodePort" --port 8080 --selector='run=forum-webserver'
kubectl expose deployment/image-webserver --type="NodePort" --port 80 --selector='run=image-webserver'

kubectl get services
kubectl get pods
kubectl get services/forum-webserver -o go-template='{{(index .spec.ports 0).nodePort}}'

# Deploy FortiWeb
# Use fortinet-solutions-cse/testbeds/fortiweb
wget https://raw.githubusercontent.com/fortinet-solutions-cse/testbeds/master/fortiweb/start_fwb_k8s.sh
chmod +x ./start_fwb_k8s.sh
./start_fwb_k8s.sh <fortiweb_kvm_qcow2_image>

# Change http port in FortiWeb to other than 80

# Create an ingress resource in K8S
kubectl apply -f ingress_example_demo.yaml
kubectl get ingress
kubectl describe ingress

# Access pods
kubectl exec -it forum-webserver-d4f956cbc-v88lz bash
