package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"context"
)

func main() {
	cli := cobra.Command{
		Use:   "kj",
		Short: "Get jobs from a namespace",
		Run: func(cmd *cobra.Command, args []string) {
			k, err := cmd.Flags().GetString("kubeconfig")
			if err != nil {
				log.Fatal(err)
			}

			var configFile string
			if k == "" {
				home := homedir.HomeDir()
				configFile = filepath.Join(home, ".kube", "config")
			} else {
				configFile = k
			}

			kubeconfig, err := clientcmd.BuildConfigFromFlags("", configFile)
			if err != nil {
				panic(err.Error())
			}

			clientset, err := kubernetes.NewForConfig(kubeconfig)
			if err != nil {
				panic(err.Error())
			}


			ns, err := cmd.Flags().GetString("namespace")
			if err != nil {
				log.Fatal(err)
			}

			if !ensureNamespace(ns, clientset) {
				log.Fatal(fmt.Printf("Namespace %s not found", ns))
			}


			jobs, err := clientset.BatchV1().Jobs(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}

			for _, j := range jobs.Items {
				if isValidJob(j) {
					fmt.Println(transformJob(j))
				}
			}
		},
	}

	cli.PersistentFlags().StringP("kubeconfig", "", "", "Path to kubeconfig file")
	cli.PersistentFlags().StringP("namespace", "n", "default", "Namespace to get jobs from")

	cli.Execute()
}

func ensureNamespace(namespace string, clientset *kubernetes.Clientset) bool {
	ns, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, n := range ns.Items {
		if n.Name == namespace {
			return true
		}
	}

	return false
}

// Completed successfully
func isValidJob(job batchv1.Job) bool {
	// job.Status.Conditions[0].Status
	containers := len(job.Spec.Template.Spec.Containers)

	isCompleted := false
	if !job.Status.CompletionTime.IsZero() {
		isCompleted = true
	}

	if isCompleted && job.Status.Succeeded == int32(containers) {
		return true
	} else {
		return false
	}
}

// transformJob returns a CSV-row format of a job's information
// this function presume job is completed
func transformJob(job batchv1.Job) string {
	job_id := job.Name
	start_time := job.Status.StartTime.Time
	duration := job.Status.CompletionTime.Sub(start_time)

	return fmt.Sprintf("%s,%s,%s", job_id, start_time, duration)
}
