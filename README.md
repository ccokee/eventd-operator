# EventD Kubernetes Operator

The **EventD Operator** is a Kubernetes operator that watches for events happening in a Kubernetes cluster and provides two main functionalities:

1. Sends event notifications to a specified Telegram channel using a Telegram bot, filtering messages by type (e.g., "Error," "Warning").

2. Publishes Kubernetes events filtered to Google Cloud Pub/Sub topics using a custom resource called "Publisher."

The operator is implemented using the **Controller Runtime** framework, which provides a set of libraries and utilities for building Kubernetes controllers. It leverages the **Kubernetes client-go** library to interact with the Kubernetes API and monitor events.

## Project Structure

The project follows a typical Go project structure. The important files and directories are as follows:

- `api/`: Contains the API definitions for custom resources used by the operator.
  - `v1alpha1/`: Contains the custom resource definitions for "Watcher" and "Publisher."
- `config/`: Contains the Kubernetes configurations for deploying the operator.
- `controllers/`: Contains the main implementation of the operator logic.
  - `watcher_controller.go`: Defines the WatcherReconciler that reconciles Watcher objects.
  - `publisher_controller.go`: Defines the PublisherReconciler that reconciles Publisher objects.
- `main.go`: The main entry point of the operator.
- `go.mod` and `go.sum`: Files that specify the Go module and its dependencies.

## How It Works

### Telegram Notification Functionality

The operator works by reconciling `Watcher` objects, which represent the desired state of the system for Telegram notifications. When a `Watcher` object is created, the operator sets up event monitoring for the specified namespace or the entire cluster. It then processes the events, retrieves their details, and sends notifications to the specified Telegram channel using the provided Telegram bot.

### Google Cloud Pub/Sub Functionality

The operator also supports the publishing of Kubernetes events to Google Cloud Pub/Sub topics. This functionality is achieved through the custom resource "Publisher."

When a `Publisher` object is created, the operator ensures the proper configuration is provided, including the GCP service account key, project ID, Pub/Sub topic, and allowed message types. The operator then processes Kubernetes events, filters them based on allowed message types, and publishes the event data to the specified GCP Pub/Sub topic.

## Prerequisites

To run the EventD Operator, you need the following prerequisites:

- A running Kubernetes cluster.
- Access to create custom resources and deploy controllers in the cluster.
- A Telegram bot and access to a Telegram channel for notifications.
- Google Cloud Platform (GCP) credentials with the necessary permissions to publish messages to Pub/Sub topics.

## Deployment

To deploy the EventD Operator, follow these steps:

1. Configure the `config/manager/kustomization.yaml` file with the appropriate values for your environment.
2. Apply the Kubernetes configuration by running `kubectl apply -k config/default`.
3. Verify that the operator is running by checking the logs of the operator pod.

## Customization

You can customize the behavior of the operator by modifying the `Watcher` and `Publisher` custom resource definitions (CRDs). The CRDs allow you to specify various settings, such as Telegram bot token, Telegram channel ID, allowed message types, GCP service account key, GCP project ID, Pub/Sub topic, and more.

## Example Watcher CRD

```yaml
apiVersion: eventd.redrvm.cloud/v1alpha1
kind: Watcher
metadata:
  name: sample-watcher
spec:
  botToken: <telegram-bot-token>
  channelID: <telegram-channel-id>
  allowedMessageTypes: ["Error", "Warning"]
  namespace: all
```

## Example Publisher CRD

```yaml
apiVersion: eventd.redrvm.cloud/v1alpha1
kind: Publisher
metadata:
  name: sample-publisher
spec:
  gcpsaKey: <base64-encoded-gcp-service-account-key>
  projectID: <gcp-project-id>
  topic: <gcp-pubsub-topic>
  allowedMessageTypes: ["Error", "Warning"]
```

## Conclusion

The EventD Operator provides a convenient way to monitor events in a Kubernetes cluster and send notifications to a Telegram channel. Additionally, it offers the capability to publish filtered Kubernetes events to Google Cloud Pub/Sub topics. The project demonstrates the usage of the Controller Runtime framework to build Kubernetes operators and showcases integration with external services like Telegram and GCP Pub/Sub.

Author: Jorge Leopoldo Curbera Rodriguez

MIT License

Copyright (c) [2023] [Jorge Leopoldo Curbera Rodriguez]

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