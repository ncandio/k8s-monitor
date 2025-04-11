package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace := flag.String("namespace", "default", "namespace to watch")
	resourceType := flag.String("resource", "deployments", "resource to watch (pods, deployments, services, etc.)")
	watch := flag.Bool("watch", false, "watch resources in real time")
	interval := flag.Int("interval", 5, "interval in seconds for watching resources")

	flag.Parse()

	// Create the client configuration
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	// Get and display resources based on type
	for {
		switch *resourceType {
		case "pods", "pod":
			listPods(ctx, clientset, *namespace)
		case "deployments", "deployment":
			listDeployments(ctx, clientset, *namespace)
		case "services", "service":
			listServices(ctx, clientset, *namespace)
		case "configmaps", "configmap":
			listConfigMaps(ctx, clientset, *namespace)
		case "secrets", "secret":
			listSecrets(ctx, clientset, *namespace)
		case "nodes", "node":
			listNodes(ctx, clientset)
		default:
			fmt.Printf("Unsupported resource type: %s\n", *resourceType)
			os.Exit(1)
		}

		// If watch mode is not enabled, break after the first iteration
		if !*watch {
			break
		}

		// Clear the screen for watch mode
		fmt.Print("\033[H\033[2J")
		fmt.Printf("Watching %s in namespace %s (Ctrl+C to exit)...\n", *resourceType, *namespace)

		// Sleep for the specified interval
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func listPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-20s %-15s %-10s %-10s\n", "NAME", "STATUS", "READY", "RESTARTS", "AGE")
	for _, pod := range pods.Items {
		containerReady := fmt.Sprintf("%d/%d", getReadyContainers(pod.Status.ContainerStatuses), len(pod.Spec.Containers))
		age := formatAge(pod.CreationTimestamp.Time)
		restarts := getTotalRestarts(pod.Status.ContainerStatuses)

		fmt.Printf("%-40s %-20s %-15s %-10d %-10s\n",
			pod.Name,
			string(pod.Status.Phase),
			containerReady,
			restarts,
			age)
	}

	fmt.Printf("\nTotal pods: %d\n", len(pods.Items))
}

func listDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-10s %-10s %-10s %-10s\n", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")
	for _, deployment := range deployments.Items {
		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		age := formatAge(deployment.CreationTimestamp.Time)

		fmt.Printf("%-40s %-10s %-10d %-10d %-10s\n",
			deployment.Name,
			ready,
			deployment.Status.UpdatedReplicas,
			deployment.Status.AvailableReplicas,
			age)
	}

	fmt.Printf("\nTotal deployments: %d\n", len(deployments.Items))
}

func listServices(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-20s %-20s %-15s %-10s\n", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "AGE")
	for _, svc := range services.Items {
		externalIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
			if externalIP == "" && svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		age := formatAge(svc.CreationTimestamp.Time)

		fmt.Printf("%-40s %-20s %-20s %-15s %-10s\n",
			svc.Name,
			string(svc.Spec.Type),
			svc.Spec.ClusterIP,
			externalIP,
			age)
	}

	fmt.Printf("\nTotal services: %d\n", len(services.Items))
}

func listConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-15s %-10s\n", "NAME", "DATA", "AGE")
	for _, cm := range configMaps.Items {
		age := formatAge(cm.CreationTimestamp.Time)

		fmt.Printf("%-40s %-15d %-10s\n",
			cm.Name,
			len(cm.Data),
			age)
	}

	fmt.Printf("\nTotal configmaps: %d\n", len(configMaps.Items))
}

func listSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-15s %-15s %-10s\n", "NAME", "TYPE", "DATA", "AGE")
	for _, secret := range secrets.Items {
		age := formatAge(secret.CreationTimestamp.Time)

		fmt.Printf("%-40s %-15s %-15d %-10s\n",
			secret.Name,
			string(secret.Type),
			len(secret.Data),
			age)
	}

	fmt.Printf("\nTotal secrets: %d\n", len(secrets.Items))
}

func listNodes(ctx context.Context, clientset *kubernetes.Clientset) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("\n%-40s %-15s %-15s %-20s %-10s\n", "NAME", "STATUS", "ROLES", "VERSION", "AGE")
	for _, node := range nodes.Items {
		status := "Ready"
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status != "True" {
					status = "NotReady"
				}
				break
			}
		}

		roles := "<none>"
		if val, ok := node.Labels["kubernetes.io/role"]; ok {
			roles = val
		} else if val, ok := node.Labels["node-role.kubernetes.io/master"]; ok && val == "true" {
			roles = "master"
		} else if val, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok && val == "true" {
			roles = "control-plane"
		}

		version := node.Status.NodeInfo.KubeletVersion
		age := formatAge(node.CreationTimestamp.Time)

		fmt.Printf("%-40s %-15s %-15s %-20s %-10s\n",
			node.Name,
			status,
			roles,
			version,
			age)
	}

	fmt.Printf("\nTotal nodes: %d\n", len(nodes.Items))
}

// Helper functions
func getReadyContainers(statuses []corev1.ContainerStatus) int {
	ready := 0
	for _, status := range statuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}

func getTotalRestarts(statuses []corev1.ContainerStatus) int {
	restarts := 0
	for _, status := range statuses {
		restarts += int(status.RestartCount)
	}
	return restarts
}

func formatAge(t time.Time) string {
	duration := time.Since(t)
	if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if duration.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else if duration.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
}

func handleError(err error) {
	if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error: %v\n", statusError.ErrStatus.Message)
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}
