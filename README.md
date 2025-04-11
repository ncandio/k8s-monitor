<img src="images/OIG3.jpg" alt="k8s monitor" width="180" height="110">


# k8s-monitor
plugin for monitoring a Kubernetes cluster 


## Features

- Watch various Kubernetes resources (pods, deployments, services, configmaps, secrets, nodes)
- Filter resources by namespace
- Real-time watching with customizable refresh intervals
- Clean, tabular output format similar to `kubectl get`

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/k8s-monitor.git

# Build the binary
cd k8s-monitor
go build -o k8s-monitor
```

## Usage

```bash
# Basic usage (displays pods in default namespace)
./k8s-monitor

# Watch deployments in a specific namespace
./k8s-monitor --resource deployments --namespace kube-system

# Watch services with a 3-second refresh interval
./k8s-monitor --resource services --watch --interval 3

# Watch nodes (cluster-wide resource)
./k8s-monitor --resource nodes
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |
| `--namespace` | Namespace to watch | `default` |
| `--resource` | Resource type to watch (pods, deployments, services, configmaps, secrets, nodes) | `deployments` |
| `--watch` | Enable watch mode with automatic refresh | `false` |
| `--interval` | Refresh interval in seconds (for watch mode) | `5` |

## Requirements

- Go 1.16+
- Kubernetes cluster access
- Valid kubeconfig file

## Testing Locally

You can test this monitor locally using [Kind](https://kind.sigs.k8s.io/), a tool for running local Kubernetes clusters using Docker containers:

```bash
# Install Kind
# On Linux:
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/

# Create a cluster
kind create cluster --name test-cluster

# Use k8s-monitor with the local cluster
./k8s-monitor --resource pods
```

## License

MIT
