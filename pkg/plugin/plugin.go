package plugin

import (
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetNamespaces: return list of namespaces in the cluster
func GetNamespaces(clientset *kubernetes.Clientset, outputCh chan NamespaceResult) {
	defer close(outputCh)

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		outputCh <- NamespaceResult{v1.Namespace{}, fmt.Errorf("failed to list namespaces: %w", err)}
	}

	for _, namespace := range namespaces.Items {
		outputCh <- NamespaceResult{namespace, nil}
	}
}

// GetPods: return list of pods in the cluster
func GetPods(clientset *kubernetes.Clientset, namespace string, outputCh chan PodResult) {
	defer close(outputCh)

	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		outputCh <- PodResult{v1.Pod{}, fmt.Errorf("failed to list pods: %w", err)}
	}

	for _, pod := range pods.Items {
		outputCh <- PodResult{pod, nil}
	}
}

// GetPodLogs: return logs from the provided pod
func GetPodLogs(clientset *kubernetes.Clientset, grep, container, pod, namespace string, follow, onlyMatch bool) error {
	count := int64(250)
	podLogOpts := v1.PodLogOptions{
		Container: container,
		Follow:    follow,
		TailLines: &count,
	}

	podLogReq := clientset.CoreV1().Pods(namespace).GetLogs(pod, &podLogOpts)

	stream, err := podLogReq.Stream()
	if err != nil {
		return err
	}
	defer stream.Close()

	// Variables needed for processing when grep value is provided
	grepWithColor := "\033[32m" + grep + "\033[0m"
	lastLine := ""
	lastMatching := false

	for {
		buf := make([]byte, 20000)
		numBytes, err := stream.Read(buf)
		if numBytes == 0 {
			if !follow {
				break
			}
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		message := string(buf[:numBytes])

		switch len(grep) == 0 {
		case true:
			fmt.Print(message)

		case false:
			// Split into individual lines
			split := strings.Split(message, "\n")

			// Process when only matching lines need to be displayed
			if onlyMatch {
				for i, line := range split {
					// Change color of the matching part of the line
					line = strings.ReplaceAll(line, grep, grepWithColor)

					// If it is the first string, check if it is a continuation of the previous last line
					if i == 0 && lastMatching {
						fmt.Print(line)
						lastMatching = false
					}

					// If a matching line is found...
					if strings.Contains(line, grep) {
						// If this is the 1st line, check if is a continuation of the previous one
						if i == 0 && len(lastLine) > 0 && lastLine[len(lastLine)-1] != '\n' {
							fmt.Print(lastLine)
						}

						// Print current line
						fmt.Print(line)

						// Add new line character if it is not the last line
						if i != len(split)-1 {
							fmt.Print("\n")
						}

						// If this is the last line and not the entire log, indicate for the rest of
						// the log message to be printed
						if i == len(split)-1 && line[len(line)-1] != '\n' {
							lastMatching = true
						}
					}

					// Capture last line in case it is not the complete log message
					// (needed in the event that the next part of it has matching grep value)
					if i == len(split)-1 {
						lastLine = line
					}
				}

			} else {
				// Process when all lines need to be displayed
				for i, line := range split {
					// Change color of the matching part of the line
					line = strings.ReplaceAll(line, grep, grepWithColor)

					fmt.Print(line)

					// Do not add new line character after the last line
					// in case it is a continuation of a log message
					if i != len(split)-1 {
						fmt.Print("\n")
					}
				}
			}
		}
	}

	fmt.Println()
	return nil
}

// TODO: Implement DoesContainerExistInPod method
