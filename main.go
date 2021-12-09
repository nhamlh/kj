package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

			ns, err := cmd.Flags().GetString("namespace")
			if err != nil {
				log.Fatal(err)
			}

			outFile, err := cmd.Flags().GetString("output")
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
				log.Fatal(err.Error())
			}

			clientset, err := kubernetes.NewForConfig(kubeconfig)
			if err != nil {
				log.Fatal(err.Error())
			}

			if !ensureNamespace(ns, clientset) {
				log.Fatal(fmt.Sprintf("Namespace %s not found", ns))
			}

			jobs, err := clientset.BatchV1().Jobs(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Fatal(err.Error())
			}

			records := [][]string{}
			for _, j := range jobs.Items {
				if isValidJob(j) {
					records = append(records, transformJob(j))
				}
			}

			if err := saveTo(outFile, records); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	cli.PersistentFlags().StringP("kubeconfig", "k", "", "Path to kubeconfig file")
	cli.PersistentFlags().StringP("namespace", "n", "default", "Namespace to get jobs from")
	cli.PersistentFlags().StringP("output", "o", "stdout", "Save to file")

	cli.Execute()
}

func ensureNamespace(namespace string, clientset *kubernetes.Clientset) bool {
	ns, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
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
func transformJob(job batchv1.Job) []string {
	job_id := job.Name
	start_time := job.Status.StartTime.Time
	duration := job.Status.CompletionTime.Sub(start_time)

	return []string{job_id, start_time.String(), duration.String()}
}

func saveTo(outFile string, records [][]string) error {
	var f io.Writer

	if outFile == "stdout" {
		f = os.Stdout
	} else {
    	if _, err := os.Stat(outFile); !os.IsNotExist(err) {
    		return fmt.Errorf("File %s already exist.", outFile)
    	}

		var err error
    	f, err = os.Create(outFile)
    	if err != nil {
    		return fmt.Errorf("Error openning file for writing: %v", err)
    	}
	}

	w := csv.NewWriter(f)
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		return fmt.Errorf("Error writing records to file: %v", err)
	}

	return nil
}
