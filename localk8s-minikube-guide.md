# Go-Food Kubernetes Deployment & Testing Guide
This guide details the complete flow: from spinning up an empty cluster to testing advanced features like Horizontal Pod Autoscaling (HPA) and Vertical Pod Autoscaling (VPA) using the `gofood.local` domain!

## 0. Initializing the Local Environment
Before verifying the cluster, you must spin up Minikube, enable the required addons, and deploy the application.

### Start Minikube
```bash
minikube start
```

### Enable Required Addons
We need the NGINX Ingress Controller (for routing) and Metrics Server (for HPA resource tracking):
```bash
minikube addons enable ingress
minikube addons enable metrics-server
```

### Install Vertical Pod Autoscaler (VPA)
The VPA CRDs and recommenders aren't built into Minikube by default. You need to clone the official autoscaler repository and run the setup script:
```bash
git clone https://github.com/kubernetes/autoscaler.git /tmp/autoscaler
cd /tmp/autoscaler/vertical-pod-autoscaler 
./hack/vpa-up.sh
cd -
```

### Map the Domain Locally (macOS/Linux)
To allow your browser and Postman to recognize `gofood.local`, point it to your localhost:
```bash
sudo sh -c 'echo "127.0.0.1 gofood.local" >> /etc/hosts'
```

### Expose Minikube Ingress on Port 80
Because Docker Desktop on macOS creates its own isolated network, you must establish a secure tunnel to expose the cluster LoadBalancers and Ingress natively on port `80`.
Open a **new terminal window** and run this command (it will ask for your Mac password to bind to port 80):
```bash
minikube tunnel
```
*Leave this terminal open and running in the background!*

### Apply Kubernetes Manifests
Now apply your configuration, stateful infrastructure (Kafka, DBs), microservices, HPA, VPA, and Ingress routing rules:
```bash
kubectl apply -f k8s/
```

---

## 1. Verify Cluster State
Ensure all your pods and services are healthy:
```bash
kubectl get pods
kubectl get svc
```
*Wait a few moments if Kafka or the microservices show as `CrashLoopBackOff` or `ContainerCreating`—they take a minute to fully initialize components like Zookeeper and Postgres.*

## 2. Testing Ingress & Load Balancing
With `minikube tunnel` running and your [/etc/hosts](file:///etc/hosts) mapped, you can now hit your microservices beautifully from the browser, Postman, or curl exactly like a production domain!

Send a request to the API Gateway through the Ingress:
```bash
curl -v http://gofood.local/health
```

Or hit a specific microservice directly via the configured routing rules:
```bash
curl -v http://gofood.local/api/users/health
```

## 3. Testing Horizontal Pod Autoscaling (HPA)
The `order-service` is configured to scale horizontally when average CPU utilization exceeds `50%`.

1. **Watch the HPA status** (in terminal 1):
```bash
kubectl get hpa order-service-hpa -w
```
2. **Watch the Pods scaling** (in terminal 2):
```bash
kubectl get pods -l app=order-service -w
```
3. **Generate High Load** (in terminal 3):
This infinite loop will flood the `order-service` and spike the CPU:
```bash
# Keep this running for a few minutes
while true; do curl -s http://gofood.local/api/orders/health > /dev/null; done
```
*Observe the HPA `TARGETS` percentage climb. Once it passes 50%, Kubernetes will automatically start spinning up new replicas of the `order-service`!*

## 4. Testing Vertical Pod Autoscaling (VPA)
The `user-service` has VPA configured. Instead of adding more pods, it adjusts the memory `requests` and cpu `limits` of the existing pod.

1. **Check the current VPA Recommendations**:
```bash
kubectl describe vpa user-service-vpa
```
2. `Recommendation:` block at the bottom, will be suggesting ideal CPU and Memory limits based on actual microservice usage.
3. **Generate Load**:
```bash
while true; do curl -s http://gofood.local/api/users/health > /dev/null; done
```
*Note: Because our VPA `updateMode` is set to `"Auto"`, if the CPU spike is significant enough and sustained for several minutes, VPA will automatically terminate the under-provisioned pod and restart it with the newly recommended, higher resource limits.*
