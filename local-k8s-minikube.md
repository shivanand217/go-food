# Kubernetes Deployment

This transforms our Docker Compose setup into fully-featured Kubernetes manifests to be run on Minikube. We'll implement advanced concepts like Horizontal Pod Autoscaling (HPA) and Ingress to simulate cloud environments (like EKS or GKE).

## Approach

We will create a `k8s/` directory and structure our declarative manifests:

### 1. Enable Minikube Features
We'll run `minikube addons enable metrics-server` (required for HPA to read CPU/Memory) and `minikube addons enable ingress` (to simulate an NGINX cloud load balancer).

### 2. Configuration & Secrets
- `k8s/configmap.yaml`: Centralized environment variables instead of hardcoding them in deployments.

### 3. Infrastructure (Stateful Workloads)
- `k8s/infrastructure.yaml`: Deployments and ClusterIP Services for PostgreSQL, MongoDB, Zookeeper, Kafka, and Redis.
*Note: For a production GCP/AWS setup, you'd typically use Managed Databases (RDS/Cloud SQL) and Managed Kafka (MSK/Confluent Cloud), but we'll run them in our cluster for this simulation.*

### 4. Microservices 
- `k8s/services.yaml`: Wait, to test HPA, our Deployments **must** have `resources.requests` defined. We will configure CPU and Memory requests/limits for all our microservices (`api-gateway`, `user-service`, `restaurant-service`, `order-service`, `delivery-service`).
- We will use the Docker Hub images we just pushed: `shivanand217/go-food-<service>:v1.0`.

### 5. Ingress & Load Balancing
- `k8s/ingress.yaml`: Instead of node ports, we will define an Ingress rule routing `http://gofood.local/` to the `api-gateway` service, simulating a real domain name hitting an Application Load Balancer.

### 6. Autoscaling (HPA & VPA)
- **Horizontal Pod Autoscaling (HPA)**: `k8s/hpa.yaml`. We will set up an HPA for the `order-service` targeting 50% CPU utilization, with a minimum of 1 pod and a maximum of 5.
- **Vertical Pod Autoscaling (VPA)**: `k8s/vpa.yaml`. We will install the VPA recommender on Minikube and configure a VPA for the `user-service`. VPA will automatically analyze the pod's resource usage over time and update its CPU/Memory requests/limits if it's struggling.

### 7. Simulation & Load Testing
We will use a tool like `hey` or a simple bash loop to flood the API Gateway with traffic:
- Watch HPA horizontally scale out new pods for `order-service` to handle the load.
- Watch VPA vertically scale up the CPU/Memory requests for the `user-service` as it gets starved for resources.
