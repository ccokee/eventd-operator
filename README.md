## EventD K8s Operator

The **EventD Operator** is a Kubernetes operator that watches for events happening in a Kubernetes cluster and sends notifications to a specified Telegram channel using a Telegram bot.

The operator is implemented using the **Controller Runtime** framework, which provides a set of libraries and utilities for building Kubernetes controllers. It leverages the **Kubernetes client-go** library to interact with the Kubernetes API and monitor events.

## Project Structure

The project follows a typical Go project structure. The important files and directories are as follows:

- `api/`: Contains the API definitions for custom resources used by the operator.
- `config/`: Contains the Kubernetes configurations for deploying the operator.
- `controllers/`: Contains the main implementation of the operator logic.
  - `watcher_controller.go`: Defines the WatcherReconciler that reconciles Watcher objects.
- `main.go`: The main entry point of the operator.
- `go.mod` and `go.sum`: Files that specify the Go module and its dependencies.

## How It Works

The operator works by reconciling `Watcher` objects, which represent the desired state of the system. It continuously monitors events in the Kubernetes cluster and compares them against the desired state specified in the `Watcher` object. If there is a mismatch, it performs the necessary operations to bring the cluster state closer to the desired state.

When a `Watcher` object is created, the operator sets up event monitoring for the specified namespace or the entire cluster.

The operator processes the events by retrieving their details, such as type, resource, description, and age. It then publishes the event messages and sends them to the specified Telegram channel using the Telegram bot.

## Prerequisites

To run the EventD Operator, you need the following prerequisites:

- A running Kubernetes cluster.
- Access to create custom resources and deploy controllers in the cluster.
- A Telegram bot and access to a Telegram channel.

## Deployment

To deploy the EventD Operator, follow these steps:

1. Configure the `config/manager/kustomization.yaml` file with the appropriate values for your environment.
2. Apply the Kubernetes configuration by running `kubectl apply -k config/default`.
3. Verify that the operator is running by checking the logs of the operator pod.

## Customization

You can customize the behavior of the operator by modifying the `Watcher` custom resource definition (CRD). The CRD allows you to specify the project ID, topic ID, bot token, channel ID, and other settings for the operator.

## Example CRD

```
apiVersion: eventd.redrvm.cloud/v1alpha1
kind: Watcher
metadata:
  name: sample-watcher
spec:
  botToken: <telegram-bot-token>
  channelID: <telegram-channel-id>
  namespace: all

```

## Conclusion

The EventD Operator provides a convenient way to monitor events in a Kubernetes cluster and send notifications to a Telegram channel. It demonstrates the usage of the Controller Runtime framework to build Kubernetes operators and showcases integration with external services.

Please refer to the project's documentation for detailed instructions on how to deploy, configure, and use the EventD Operator.

If you have any further questions or need assistance, please feel free to reach out.

MIT License

Copyright (c) [year] [project owner]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
