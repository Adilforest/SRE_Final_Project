package main

import (
  "context"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "path/filepath"
  "time"

  v1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
)

func main() {
  // Load kubeconfig from default location
  kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
  config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
  if err != nil {
    panic(err.Error())
  }

  // Create Kubernetes client
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err.Error())
  }

  // Define namespace and applications to monitor
  namespace := "default"
  apps := []string{"api-gateway", "auth", "product", "order"}

  // Create output directory with timestamp
  timestamp := time.Now().Format("20060102_150405")
  outputDir := "sre_diagnostics_" + timestamp
  err = os.Mkdir(outputDir, 0755)
  if err != nil {
    panic(err)
  }

  // Collect logs from pods by app label
  for _, app := range apps {
    fmt.Printf("Collecting logs for %s...\n", app)

    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
      LabelSelector: "app=" + app,
    })
    if err != nil {
      fmt.Printf("Error listing pods for %s: %v\n", app, err)
      continue
    }

    for _, pod := range pods.Items {
      logs, err := getPodLogs(clientset, namespace, pod.Name)
      if err != nil {
        fmt.Printf("Error retrieving logs for pod %s: %v\n", pod.Name, err)
        continue
      }

      logFile := filepath.Join(outputDir, fmt.Sprintf("%s_%s.log", app, pod.Name))
      err = ioutil.WriteFile(logFile, []byte(logs), 0644)
      if err != nil {
        fmt.Printf("Error writing logs to file %s: %v\n", logFile, err)
      }
    }
  }

  // Collect Prometheus metrics from API Gateway
  promURL := "http://localhost:8080/metrics"
  fmt.Println("Collecting Prometheus metrics from API Gateway...")
  resp, err := http.Get(promURL)
  if err != nil {
    fmt.Printf("Error fetching Prometheus metrics: %v\n", err)
  } else {
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    metricsFile := filepath.Join(outputDir, "api-gateway_metrics.prom")
    err = ioutil.WriteFile(metricsFile, body, 0644)
    if err != nil {
      fmt.Printf("Error writing Prometheus metrics file: %v\n", err)
    }
  }

  fmt.Println("Diagnostics collection complete. Files saved in folder:", outputDir)
}

func getPodLogs(clientset *kubernetes.Clientset, namespace, podName string) (string, error) {
  req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{})
  podLogs, err := req.Stream(context.TODO())
  if err != nil {
    return "", err
  }
  defer podLogs.Close()

  logs, err := ioutil.ReadAll(podLogs)
  if err != nil {
    return "", err
  }

  return string(logs), nil
}