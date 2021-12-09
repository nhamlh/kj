package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"context"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get jobs from a namespace",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello world")
	},
}

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

			jobs, err := clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("There are %d jobs in the cluster\n", len(jobs.Items))
		},
	}

	cli.PersistentFlags().StringP("kubeconfig", "", "", "Path to kubeconfig file")

	cli.Execute()
}
