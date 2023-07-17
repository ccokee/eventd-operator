package controllers

import (
	"context"
	"fmt"
	"strconv"

	eventdv1alpha1 "github.com/ccokee/eventd-operator/api/v1alpha1"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	Domain   = "redrvm.cloud"
	Group    = "eventd"
	Version  = "v1alpha1"
	Kind     = "Watcher"
	Operator = "eventd-operator"
)

// WatcherReconciler reconciles a Watcher object
type WatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=watchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=watchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=eventd.redrvm.cloud,resources=watchers/finalizers,verbs=update
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
// TODO: Modify the Reconcile function to compare the state specified by
// the Watcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *WatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the Watcher resource
	watcher := &eventdv1alpha1.Watcher{}
	if err := r.Get(ctx, req.NamespacedName, watcher); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Obtener la configuración del cliente utilizando el contexto del pod
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(err, "Error al obtener la configuración del cliente")
		return reconcile.Result{}, err
	}

	// Crear un nuevo cliente de Kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Error al crear el cliente de Kubernetes")
		return reconcile.Result{}, err
	}

	// Obtener todos los eventos en el clúster o en un namespace específico
	var eventWatcher watch.Interface
	if watcher.Spec.Namespace == "all" {
		eventWatcher, err = clientset.CoreV1().Events("").Watch(ctx, metav1.ListOptions{})
	} else {
		eventWatcher, err = clientset.CoreV1().Events(watcher.Spec.Namespace).Watch(ctx, metav1.ListOptions{})
	}
	if err != nil {
		logger.Error(err, "Error al crear el watcher de eventos")
		return reconcile.Result{}, err
	}

	// Configurar el bot de Telegram
	botToken := watcher.Spec.BotToken
	channelID := watcher.Spec.ChannelID

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Error(err, "Error al crear el bot de Telegram")
		return reconcile.Result{}, err
	}

	bot.Debug = true
	// Procesar los eventos y enviarlos al canal de Telegram
	eventChannel := eventWatcher.ResultChan()
	for rawEvent := range eventChannel {
		event, ok := rawEvent.Object.(*corev1.Event)
		if !ok {
			logger.Info("No se pudo convertir el evento a tipo corev1.Event", "event", rawEvent)
			continue
		}

		// Obtener los campos del evento
		eventType := event.Type
		eventResource := event.InvolvedObject.Kind
		eventDescription := event.Message
		eventAge := event.CreationTimestamp.String()

		// Comprobar si el tipo de mensaje del evento está permitido
		if !isMessageTypeAllowed(eventType, watcher.Spec.AllowedMessageTypes) {
			// El tipo de mensaje no está permitido, omitir el envío
			continue
		}

		// Enviar el mensaje al canal de Telegram
		channelIDInt, err := strconv.ParseInt(channelID, 10, 64)
		if err != nil {
			logger.Error(err, "Error al convertir el canal ID a int64")
			continue
		}
		message := tgbotapi.NewMessage(channelIDInt, fmt.Sprintf("Type: %s\nResource: %s\nDescription: %s\nAge: %s", eventType, eventResource, eventDescription, eventAge))

		_, err = bot.Send(message)
		if err != nil {
			logger.Error(err, "Error al enviar el mensaje al canal de Telegram")
		}

		logger.Info("Evento enviado al canal de Telegram", "event", event)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventdv1alpha1.Watcher{}).
		Complete(r)
}

func isMessageTypeAllowed(messageType string, allowedMessageTypes []string) bool {
	for _, allowedType := range allowedMessageTypes {
		if allowedType == messageType {
			return true
		}
	}
	return false
}
