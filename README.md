# Kubectl Plugin: Simple kubectl plugin! 

A `kubectl` plugin to simply every everyday common operations against a Kubernetes clusters. Current functionality includes:

- Getting a list of namespaces in the cluster
- Getting a very high-level, quick status on your deployments
- Accessing log messages without actually having to type out pod names
    - Bonus: you can also filter out log messages based on a keyword/phrase!

## Installation
1. Go to the bin directory, pick the appropriate executable for your OS (for Mac use Darwin), and download it.
2. Move it to any of the locations within your `PATH` environment variable.
3. Update the permissions on the executable file to ensure that you have the permission to execute the file.
4. Run the command: `kubectl simple -h` to verify that the plugin has been successfully installed.