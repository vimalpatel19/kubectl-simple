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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vimalpatel19/kubectl-simple/pkg/logger"
	"github.com/vimalpatel19/kubectl-simple/pkg/plugin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output log messages from a pod",
	Long: `Return an output of log messages from a Kubernetes pod matching
	the provided name. If there are multiple pods matching the provided name, 
	the user can select which pod to output the log messages from. Additional
	options are available for filtering out log messages based on a keyword/phrase.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
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

		follow, _ := cmd.Flags().GetBool(FOLLOW_LOGS_PARAM)
		container, _ := cmd.Flags().GetString(CONTAINER_PARAM)
		namespace, _ := cmd.Flags().GetString(NAMESPACE_PARAM)
		phrase, _ := cmd.Flags().GetString(GREP_LOGS_PARAM)
		onlyMatching, _ := cmd.Flags().GetBool(ONLY_MATCHING_LOGS_PARAM)

		// Ensure a keyword/phrase is provided if the only-matching flag is set to true
		if onlyMatching && len(phrase) == 0 {
			log.Error(errors.New("please provide a grep keyword/phrase with the only-matching flag"))
		}

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

		var loggingPod v1.Pod
		switch len(matchingPods) == 1 {
		case true:
			// Handle when matching pod is found
			loggingPod = matchingPods[0]
			log.Info("\n\n✓ Found a matching pod: %s\n", loggingPod.Name)
			s.Stop()

			log.Info("Outputing logs...")
			callGetPodsLogs(clientset, phrase, container, loggingPod, follow, onlyMatching)

		case false:
			// Ask user to pick from one of the multiple matching pods
			var selected int
			log.Info("\n\n✓ Found matching pods:\n")
			s.Stop()

			// Output each of the matching pods
			for i, pod := range matchingPods {
				fmt.Printf("\t %d. %s\n", i+1, pod.Name)
			}

			// Ask user to select one of the matching pods and validate user input
			fmt.Printf("\nPlease select one of the options above and press enter: ")
			if _, err := fmt.Scan(&selected); err != nil {
				log.Info("")
				log.Error(errors.New("✖ please provide a valid input"))
				return
			}

			if selected < 1 || selected > len(matchingPods) {
				log.Info("")
				log.Error(errors.New("✖ please provide a valid input"))
				return
			}

			// Output log messages for the selected pod
			loggingPod = matchingPods[selected-1]
			log.Info("\nOutputing logs from %s...", loggingPod.Name)
			callGetPodsLogs(clientset, phrase, container, loggingPod, follow, onlyMatching)
		}
	},
}

// callGetPodsLogs: helper function that calls GetPodLogs method with the correct container name
func callGetPodsLogs(clientset *kubernetes.Clientset, grep, container string, pod v1.Pod, tail, onlyMatching bool) {
	// Use provider container if given, otherwise use default name (i.e. service name)
	if len(container) > 0 {
		// TODO: Add call to plugin.DoesContainerExistInPod method after it has been implemented
		plugin.GetPodLogs(clientset, grep, container, pod.Name, pod.Namespace, tail, onlyMatching)
	} else {
		splitPodName := strings.Split(pod.Name, "-")
		plugin.GetPodLogs(clientset, grep, splitPodName[0], pod.Name, pod.Namespace, tail, onlyMatching)
	}
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	logsCmd.Flags().StringP(NAMESPACE_PARAM, "n", "", "Namespace to filter results with")
	logsCmd.Flags().StringP(CONTAINER_PARAM, "c", "", "Container in the pod")
	logsCmd.Flags().BoolP(FOLLOW_LOGS_PARAM, "f", false, "Indicate whether to follow the logs")
	logsCmd.Flags().StringP(GREP_LOGS_PARAM, "g", "", "Keyword/phrase to grep from incoming logs")
	logsCmd.Flags().BoolP(ONLY_MATCHING_LOGS_PARAM, "m", false, "Indicate whether to output only matching logs")
}
