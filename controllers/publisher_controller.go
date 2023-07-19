/*
Copyright 2023.

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

package controllers

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/watch"

	eventdv1alpha1 "github.com/ccokee/eventd-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// PublisherReconciler reconciles a Publisher object
type PublisherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=publishers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=publishers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=publishers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=serviceaccounts,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Publisher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *PublisherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the Publisher resource
	publisher := &eventdv1alpha1.Publisher{}
	if err := r.Get(ctx, req.NamespacedName, publisher); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the Publisher is configured properly
	if !r.isPublisherConfigured(publisher) {
		// Set the "Configured" condition to "False" if the configuration is incomplete
		r.setPublisherCondition(publisher, eventdv1alpha1.ConditionConfigured, metav1.ConditionFalse, "IncompleteConfiguration", "Publisher configuration is incomplete.")
		if err := r.Status().Update(ctx, publisher); err != nil {
			logger.Error(err, "Error updating Publisher status")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Set the "Configured" condition to "True" if the configuration is complete
	r.setPublisherCondition(publisher, eventdv1alpha1.ConditionConfigured, metav1.ConditionTrue, "Ready", "Publisher is configured and ready to publish events.")
	if err := r.Status().Update(ctx, publisher); err != nil {
		logger.Error(err, "Error updating Publisher status")
		return reconcile.Result{}, err
	}

	// Start a separate goroutine to watch for Kubernetes events and process them
	go r.processEvents(ctx, publisher)

	return ctrl.Result{}, nil
}

func (r *PublisherReconciler) isPublisherConfigured(publisher *eventdv1alpha1.Publisher) bool {
	// Check if all necessary fields in PublisherSpec are present
	return publisher.Spec.GCPSAKey != "" &&
		publisher.Spec.ProjectID != "" &&
		publisher.Spec.Topic != "" &&
		len(publisher.Spec.AllowedMessageTypes) > 0
}

func (r *PublisherReconciler) setPublisherCondition(publisher *eventdv1alpha1.Publisher, conditionType eventdv1alpha1.PublisherConditionType, status metav1.ConditionStatus, reason, message string) {
	condition := eventdv1alpha1.PublisherCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
	publisher.Status.Conditions = append(publisher.Status.Conditions, condition)
}

func (r *PublisherReconciler) processEvents(ctx context.Context, publisher *eventdv1alpha1.Publisher) {
	logger := log.FromContext(ctx)

	// GCP authentication using the service account key stored in the Publisher CRD
	// Decode the base64-encoded GCP service account key
	serviceAccountKey, err := base64.StdEncoding.DecodeString(publisher.Spec.GCPSAKey)
	if err != nil {
		logger.Error(err, "Error decoding GCP service account key")
		return
	}

	// Create a new Pub/Sub client with the decoded service account key
	client, err := pubsub.NewClient(ctx, publisher.Spec.ProjectID, option.WithCredentialsJSON(serviceAccountKey))
	if err != nil {
		logger.Error(err, "Error creating Pub/Sub client")
		return
	}

	// Create a Pub/Sub topic handle
	topic := client.Topic(publisher.Spec.Topic)

	// Watch Kubernetes events based on the AllowedMessageTypes list
	eventWatcher, err := r.watchEvents(ctx, publisher)
	if err != nil {
		// Handle the error
		return
	}

	// Continuously publish events to the specified GCP Pub/Sub topic
	for rawEvent := range eventWatcher.ResultChan() {
		event, ok := rawEvent.Object.(*corev1.Event)
		if !ok {
			logger.Info("Failed to convert event to corev1.Event", "event", rawEvent)
			continue
		}

		// Check if the event type is allowed
		if isMessageTypeAllowed(event.Type, publisher.Spec.AllowedMessageTypes) {
			// Publish the event data to the Pub/Sub topic
			res := topic.Publish(ctx, &pubsub.Message{
				Data: []byte(fmt.Sprintf("Type: %s, Resource: %s, Description: %s, Age: %s", event.Type, event.InvolvedObject.Kind, event.Message, event.CreationTimestamp.String())),
			})

			// Log the result (can be used for error handling if needed)
			logger.Info("Published event to Pub/Sub", "result", res)
		}
	}
}

func (r *PublisherReconciler) watchEvents(ctx context.Context, publisher *eventdv1alpha1.Publisher) (watch.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	eventWatcher, err := clientset.CoreV1().Events(publisher.Namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return eventWatcher, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PublisherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventdv1alpha1.Publisher{}).
		Complete(r)
}
