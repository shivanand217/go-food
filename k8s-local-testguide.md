# Kubernetes Local Testing Guide

This guide details how to verify and test the advanced Kubernetes features we've deployed locally using Minikube. 

## 1. Verify Cluster State
First, ensure all your pods and services are healthy:
```bash
kubectl get pods
kubectl get svc
```
*It may take a few moments for Kafka, Zookeeper, and the Go microservices to fully initialize.*

## 2. Testing Ingress & Load Balancing
On macOS, the Minikube IP is usually not directly accessible from your host machine's terminal. To fix this, we port-forward the NGINX Ingress Controller to your local machine.
1. Run this command in a **new terminal tab** and leave it running:
```bash
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80
```
2. Send a request to the API Gateway through the Ingress (in your main terminal):
```bash
curl -v -H "Host: gofood.local" http://localhost:8080/health
```

## 3. Testing Horizontal Pod Autoscaling (HPA)
The `order-service` is configured to scale horizontally when average CPU utilization exceeds 50%.
1. **Watch the HPA status** (in terminal 1):
```bash
kubectl get hpa order-service-hpa -w
```
2. **Watch the Pods scaling** (in terminal 2):
```bash
kubectl get pods -l app=order-service -w
```
3. **Generate High Load** (in terminal 3). This infinite loop will spike the CPU:
```bash
# Keep this running for a few minutes
while true; do curl -s -H "Host: gofood.local" http://localhost:8080/api/orders/health > /dev/null; done
```
*Observe the HPA CPU percentage climb. Once it passes 50%, Kubernetes will automatically start creating new replicas of the `order-service`!*

## 4. Testing Vertical Pod Autoscaling (VPA)
The `user-service` has VPA configured. Instead of adding more pods, it adjusts the `requests` and `limits` of the existing pod.
1. **Check the current VPA Recommendations**:
```bash
kubectl describe vpa user-service-vpa
```
2. Look for the `Recommendation:` block at the bottom. It will suggest ideal CPU and Memory limits based on actual usage.
3. **Generate Load**:
```bash
while true; do curl -s -H "Host: gofood.local" http://localhost:8080/api/users/health > /dev/null; done
```
*Note: Because our VPA `updateMode` is set to `"Auto"`, if the CPU spike is significant enough and sustained, VPA will automatically terminate the under-provisioned pod and start a new one with the newly recommended, higher resource limits.*

---

*I have started minikube back up for you in the background.*
