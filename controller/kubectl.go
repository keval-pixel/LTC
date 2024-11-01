package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientset() *kubernetes.Clientset {
	// Load kubeconfig
	kubeconfig := filepath.Join(
		homeDir(), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func deploy_replicas(replicas_int int32, tool string, containerimage string) string {

	clientset := getClientset()

	if checkDeploymentExists(tool) {
		fmt.Printf("A Deployment with name %s already exists!\nDelete this deployment and then run start\n", tool)
		return ""
	}

	// Define Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: tool,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas_int),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": tool,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": tool,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            tool,
							Image:           containerimage,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					NodeSelector: map[string]string{
						"stress-test": "true", // Node selector here
					},
				},
			},
		},
	}

	// Create Deployment
	deploymentsClient := clientset.AppsV1().Deployments(corev1.NamespaceDefault)
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		color.Red("Error creating a deployment: %v\n", err)
		return ""
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return result.GetObjectMeta().GetName()
}

func update_replicas(replicas int32, deploymentname string) {

	if replicas < 0 {
		color.Red("Cannot update the number of replicas with a negative number")
		return
	}

	clientset := getClientset()

	// Define the deployment name, namespace, and desired number of replicas
	deploymentName := deploymentname
	namespace := corev1.NamespaceDefault

	// Get the current deployment
	deployment := get_deployment(deploymentName)
	if deployment == nil {
		return
	}

	pending_replicas := *deployment.Spec.Replicas - deployment.Status.ReadyReplicas

	if deployment.Status.ReadyReplicas >= pending_replicas || replicas < *deployment.Spec.Replicas {
		// Update the number of replicas
		deployment.Spec.Replicas = &replicas

		// Apply the updated deployment
		_, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, v1.UpdateOptions{})
		if err != nil {
			color.Red("Error updating deployment: %s", err.Error())
		} else {
			color.Green("Successfully scaled deployment %s to %d replicas\n", deploymentName, replicas)
		}

	} else {
		color.Yellow("Warning: Unable to update replicas to %d. Pending replicas to be deployed are: %d\n", replicas, pending_replicas)
		color.White("Number of replicas running is %d\n", deployment.Status.ReadyReplicas)
	}
}

func delete_deployment(deploymentname string) error {
	clientset := getClientset()

	// Define the deployment name, namespace, and desired number of replicas
	deploymentName := deploymentname
	namespace := corev1.NamespaceDefault

	// Delete the deployment
	deletePolicy := metav1.DeletePropagationForeground
	err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		fmt.Printf("Error deleting deployment: %v\n", err)
		return err
	}

	color.Green("Deployment %s deleted successfully\n", deploymentName)

	return err
}

// checkDeploymentExists checks if the deployment exists in the given namespace
func checkDeploymentExists(deploymentName string) bool {
	clientset := getClientset()
	namespace := corev1.NamespaceDefault
	_, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Deployment does not exist
			return false
		}
		fmt.Printf("Error checking deployment: %v\n", err)
		// Some other error occurred
		return false
	}
	// Deployment exists
	return true
}

func get_replicas(deploymentName string) int32 {
	deployment := get_deployment(deploymentName)
	return *deployment.Spec.Replicas
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func int32Ptr(i int32) *int32 { return &i }

// Prometheus query result format
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func get_pods(deploymentName string) {
	clientset := getClientset()
	namespace := corev1.NamespaceDefault
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting deployment: %v\n", err)
		return
	}

	// Get the label selector from the deployment spec
	labelSelector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	// List Pods with the same label selector
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	currentTime := time.Now()

	// Print pod names
	fmt.Printf("Pods for deployment %s:\n\n", deploymentName)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	// Print the table header
	fmt.Fprintf(w, "POD\tSTATUS\tNODE\tRUNTIME\t\n")
	fmt.Fprintf(w, "---\t------\t----\t-------\t\n")

	nodePodCount := make(map[string]int)

	// Loop through the pods and count them per node

	for _, pod := range pods.Items {
		if pod.Status.StartTime != nil {
			nodeName := pod.Spec.NodeName

			// If the pod is assigned to a node (i.e., nodeName is not empty)
			if nodeName != "" {
				nodePodCount[nodeName]++
			}
			startTime := pod.Status.StartTime.Time
			duration := currentTime.Sub(startTime)
			roundedDuration := duration.Round(time.Second)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", pod.Name, pod.Status.Phase, pod.Spec.NodeName, roundedDuration)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\tNot started\t\n", pod.Name, pod.Status.Phase, pod.Spec.NodeName)
		}
	}

	w.Flush()

	fmt.Print("\n\n")
	for nodeName, count := range nodePodCount {
		color.Yellow("Node: %s, Pod Count: %d\n", nodeName, count)
	}
}

func get_deployment(deploymentName string) *appsv1.Deployment {
	clientset := getClientset()
	namespace := corev1.NamespaceDefault
	// Use the AppsV1 API to get the deployment
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})

	if err != nil {
		color.Red("Error getting deployment: %v", err)
		return nil
	}

	return deployment
}

func update_deployment(deployment *appsv1.Deployment) {
	clientset := getClientset()
	namespace := corev1.NamespaceDefault
	_, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, v1.UpdateOptions{})
	if err != nil {
		log.Fatalf("Error updating deployment: %s", err.Error())
	}
}

func get_pod_logs(podName string) {

	clientset := getClientset()

	// Define namespace and pod name
	namespace := "default" // Replace with your namespace

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling to catch SIGINT or SIGTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt signal, stopping log stream...")
		cancel()
	}()

	// Start streaming logs
	streamLogs(ctx, clientset, namespace, podName)
}

// streamLogs streams logs from a specific pod and container, following them until context is canceled
func streamLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string) {
	// Prepare the log options (similar to --follow)
	logOptions := &corev1.PodLogOptions{
		Follow: true,
	}

	// Request the logs for the specified pod
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)

	// Get a stream to the logs
	stream, err := req.Stream(ctx)
	if err != nil {
		fmt.Printf("Error opening log stream: %v\n", err)
		return
	}
	defer stream.Close()

	// Read the log stream line by line
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		// Check if the context has been canceled
		select {
		case <-ctx.Done():
			fmt.Println("Log stream stopped.")
			return
		default:
			fmt.Println(scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log stream: %v\n", err)
	}
}

func label_nodes() {
	clientset := getClientset()

	labelSelector := "stress-test=true"
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Fatalf("Error listing nodes: %v", err)
	}

	if len(nodes.Items) == 0 {
		fmt.Println("No nodes found with the label:", labelSelector)
	} else {
		fmt.Printf("Nodes with label %s:\n", labelSelector)
		for _, node := range nodes.Items {
			cpu := node.Status.Capacity[corev1.ResourceCPU]
			node.Labels["cpu-cores"] = cpu.String()
		}
	}
}

func delete_label_nodes() {
	clientset := getClientset()

	labelSelector := "stress-test=true"
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Fatalf("Error listing nodes: %v", err)
	}

	if len(nodes.Items) == 0 {
		fmt.Println("No nodes found with the label:", labelSelector)
	} else {
		for _, node := range nodes.Items {
			delete(node.Labels, "cpu-cores")
			_, err := clientset.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{})
			if err != nil {
				log.Printf("Error updating node %s: %v", node.Name, err)
			}
		}
	}
}
