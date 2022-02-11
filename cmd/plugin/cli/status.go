/*
Copyright © 2022 VIMAL PATEL vimalpatel0611@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vimalpatel19/kubectl-simple/pkg/logger"
	"github.com/vimalpatel19/kubectl-simple/pkg/plugin"
	v1 "k8s.io/api/core/v1"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of deployment",
	Long: `Return the status of all deployments matching the provided name by 
	returning the status of each deployment's pods. Results can be filtered out
	by providing a target namespace.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()

		// Validate and get input arguments/parameters
		inputName := ""
		switch len(args) {
		case 0:
			log.Error(errors.New("please provide the name of a deployment/service"))
			return
		case 1:
			inputName = args[0]
		default:
			log.Error(errors.New("please provide the name of only one deployment/service"))
			return
		}

		namespace, _ := cmd.Flags().GetString(NAMESPACE_PARAM)

		// Get a Kubernetes clientset
		clientset, err := plugin.GetClientset(KubernetesConfigFlags)
		if err != nil {
			log.Error(err)
			return
		}

		// Make a call to get list of pods
		log.Info("")
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Prefix = "Searching for matching pod(s) "
		s.Start()

		podCh := make(chan plugin.PodResult, 1)
		matchingPods := []v1.Pod{}

		go plugin.GetPods(clientset, namespace, podCh)

		for result := range podCh {
			if result.Err != nil {
				log.Error(result.Err)
				return
			} else {
				if strings.Contains(result.Value.Name, inputName) {
					matchingPods = append(matchingPods, result.Value)
				}
			}
		}

		// Return if no matching pods were found
		if len(matchingPods) == 0 {
			log.Error(fmt.Errorf("\n\n✖ no pods found matching the provided name"))
			s.Stop()
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(matchingPods))

		log.Info("\n\n✓ Found matching pods\n")
		for _, pod := range matchingPods {
			// Asynchronously output the status of each pod
			go func(p v1.Pod) {
				// Check status of current pod
				for _, condition := range p.Status.Conditions {
					if condition.Type == v1.PodReady {
						switch condition.Status {
						case v1.ConditionTrue:
							printPodStatus(p, "Ready", color.FgHiGreen)
						case v1.ConditionFalse:
							printPodStatus(p, "Not ready", color.FgHiRed)
						case v1.ConditionUnknown:
							printPodStatus(p, "Unknown", color.FgHiYellow)
						}
					}
				}
				wg.Done()

			}(pod)
		}

		wg.Wait()
	},
}

// printPodStatus: helper method that generates the output to print for the given pod
func printPodStatus(pod v1.Pod, status string, c color.Attribute) {
	var b bytes.Buffer
	printColor := color.New(c)

	fmt.Fprintf(&b, "POD %s: %s\n", pod.Name, status)
	fmt.Fprintf(&b, "  CONTAINERS\n")

	for _, c := range pod.Status.ContainerStatuses {
		fmt.Fprintf(&b, "  %-14s Ready: %-5t Restarts: %d\n", c.Name, c.Ready, c.RestartCount)
	}

	printColor.Println(b.String())
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	statusCmd.Flags().StringP(NAMESPACE_PARAM, "n", "", "Namespace to filter results with")
}
