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
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vimalpatel19/kubectl-simple/pkg/logger"
	"github.com/vimalpatel19/kubectl-simple/pkg/plugin"
)

// namespacesCmd represents the namespaces command
var namespacesCmd = &cobra.Command{
	Use:           "namespaces",
	Short:         "List of namespaces",
	Long:          `Return a list of all the namespaces found in the Kubernetes cluster`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		namespaceCh := make(chan plugin.NamespaceResult, 1)

		// Get a Kubernetes clientset
		clientset, err := plugin.GetClientset(KubernetesConfigFlags)
		if err != nil {
			log.Error(err)
			return
		}

		// Make call to get list of namespaces
		go plugin.GetNamespaces(clientset, namespaceCh)

		// Start cli spinner while waiting for namespaces to be received
		log.Info("")
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Prefix = "Searching for namespaces "
		s.Start()

		// Updated console output as each namespace is received
		for result := range namespaceCh {
			if result.Err != nil {
				s.Stop()
				log.Error(err)
				return
			} else {
				s.Restart()
				log.Info("%s", result.Value.Name)
			}
		}

		// Wrap up on the cli spinner
		s.FinalMSG = "\nSearching for namespaces completed ✓\n"
		s.Stop()
	},
}

func init() {
	rootCmd.AddCommand(namespacesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// namespacesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// namespacesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
