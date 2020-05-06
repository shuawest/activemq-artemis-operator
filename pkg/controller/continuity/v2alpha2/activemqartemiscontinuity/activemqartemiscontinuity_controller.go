package activemqartemiscontinuity

import (
	"context"
	"strconv"
	"time"

	continuityv2alpha2 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/continuity"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/management/jolokia"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/secrets"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_activemqartemiscontinuity")

// Add creates a new ActiveMQArtemisContinuity Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileActiveMQArtemisContinuity{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("activemqartemiscontinuity-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemisContinuity
	err = c.Watch(&source.Kind{Type: &continuityv2alpha2.ActiveMQArtemisContinuity{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemisContinuity
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &continuityv2alpha2.ActiveMQArtemisContinuity{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileActiveMQArtemisContinuity implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileActiveMQArtemisContinuity{}

// ReconcileActiveMQArtemisContinuity reconciles a ActiveMQArtemisContinuity object
type ReconcileActiveMQArtemisContinuity struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ActiveMQArtemisContinuity object and makes changes based on the state read
// and what is in the ActiveMQArtemisContinuity.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileActiveMQArtemisContinuity) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemisContinuity")

	// Fetch the ActiveMQArtemisContinuity instance
	instance := &continuityv2alpha2.ActiveMQArtemisContinuity{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("IsNotFound no requeue")
			return reconcile.Result{}, nil

			// TODO: should continuity be shut down when CR is removed?
		}

		// Error reading the object - requeue the request.
		reqLogger.Info("Requeuing due to continuity err: " + err.Error())
		return reconcile.Result{}, err
	} else {
		// Found continuity CR - boot continuity
		err = bootContinuity(instance, request, r.client)
		if nil != err {
			// Error calling continuity - requeue the request
			reqLogger.Info("Error while booting continuity: " + err.Error())
			return reconcile.Result{}, err
		} else {
			isBooted, err := retrieveBooted(instance, request, r.client)
			if nil != err {
				// Error calling verifying continuity status - requeue the request
				reqLogger.Info("Error verifying continuity status: " + err.Error())
				return reconcile.Result{}, err
			}

			if !isBooted {
				reqLogger.Info("Contintuity not booted yet")
				return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
			} else {
				reqLogger.Info("Contintuity is booted")
				return reconcile.Result{}, nil
			}
		}
	}

	return reconcile.Result{}, nil
}

func retrieveBooted(instance *continuityv2alpha2.ActiveMQArtemisContinuity, request reconcile.Request, client client.Client) (bool, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("ActiveMQArtemisContinuity checking bootstrap status")

	isBooted := true
	var err error = nil

	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for i, a := range artemisArray {
			if nil == a {
				isBooted = false
				continue
			}

			bootedResult, err := a.IsBooted()
			if nil != err {
				reqLogger.Info("IsBooted ActiveMQArtemisContinuity error", "Error", err)
				isBooted = false
				continue
			}

			if !bootedResult {
				isBooted = false
				continue
			}

			reqLogger.Info("ActiveMQArtemisContinuity on broker instance " + strconv.Itoa(i) + " is booted")
		}
	}

	return isBooted, err
}

func bootContinuity(instance *continuityv2alpha2.ActiveMQArtemisContinuity, request reconcile.Request, client client.Client) error {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Creating ActiveMQArtemisContinuity createContinuity")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)

	if nil != artemisArray {
		for i, a := range artemisArray {
			if nil == a {
				continue
			}

			isBooted, err := a.IsBooted()
			if nil != err {
				reqLogger.Info("Error checking if continuity (" + strconv.Itoa(i) + ") is already booted: " + err.Error())
				continue
			} else if isBooted {
				reqLogger.Info("Continuity (" + strconv.Itoa(i) + ") is already booted")
				continue
			}

			err = a.Configure(instance.Spec.SiteId, instance.Spec.ActiveOnStart, instance.Spec.ServingAcceptors, "continuity-local", instance.Spec.RemoteConnectorRefs, instance.Spec.ReorgManagement)
			if nil != err {
				reqLogger.Info("Error configuring continuity (" + strconv.Itoa(i) + "): " + err.Error())
				continue
			}

			err = a.SetSecrets(instance.Spec.LocalContinuityUser, instance.Spec.LocalContinuityPass, instance.Spec.RemoteContinuityUser, instance.Spec.RemoteContinuityPass)
			if nil != err {
				reqLogger.Info("Error setting continuity (" + strconv.Itoa(i) + ") secrets: " + err.Error())
				continue
			}

			err = a.Tune(instance.Spec.ActivationTimeout, instance.Spec.InflowStagingDelay, instance.Spec.BridgeInterval, instance.Spec.BridgeIntervalMultiplier, instance.Spec.PollDuration)
			if nil != err {
				reqLogger.Info("Error setting continuity (" + strconv.Itoa(i) + ") tuning: " + err.Error())
				continue
			}

			err = a.Boot()
			if nil != err {
				reqLogger.Info("Error booting continuity (" + strconv.Itoa(i) + "): " + err.Error())
				continue
			}

			reqLogger.Info("Booted continuity (" + strconv.Itoa(i) + ")")
		}
	}

	return err
}

func getPodBrokers(instance *continuityv2alpha2.ActiveMQArtemisContinuity, request reconcile.Request, client client.Client) []*continuity.ArtemisContinuity {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Getting Pod Brokers")

	var artemisArray []*continuity.ArtemisContinuity = nil
	var err error = nil

	ss.NameBuilder.Name()
	if err != nil {
		reqLogger.Error(err, "Failed to ge the statefulset name")
	}

	// Check to see if the statefulset already exists
	ssNamespacedName := types.NamespacedName{
		Name:      ss.NameBuilder.Name(),
		Namespace: request.Namespace,
	}

	statefulset, err := ss.RetrieveStatefulSet(ss.NameBuilder.Name(), ssNamespacedName, client)
	if nil != err {
		reqLogger.Info("Statefulset: " + ssNamespacedName.Name + " not found")
	} else {
		reqLogger.Info("Statefulset: " + ssNamespacedName.Name + " found")

		pod := &corev1.Pod{}
		podNamespacedName := types.NamespacedName{
			Name:      statefulset.Name + "-0",
			Namespace: request.Namespace,
		}

		user, pass, _ := retrieveBrokerCredentials(podNamespacedName.Namespace, client)

		// For each of the replicas
		var i int = 0
		var replicas int = int(*statefulset.Spec.Replicas)
		artemisArray = make([]*continuity.ArtemisContinuity, 0, replicas)
		for i = 0; i < replicas; i++ {
			s := statefulset.Name + "-" + strconv.Itoa(i)
			podNamespacedName.Name = s

			var artemis *continuity.ArtemisContinuity = nil

			if err = client.Get(context.TODO(), podNamespacedName, pod); err != nil {
				if errors.IsNotFound(err) {
					reqLogger.Error(err, "Pod IsNotFound")
				} else {
					reqLogger.Error(err, "Pod lookup error")
				}
			} else {
				if "" == pod.Status.PodIP {
					reqLogger.Info("Pod IP not available yet")
				} else {
					reqLogger.Info("Pod found", "PodIP", pod.Status.PodIP)

					jolokiaClient := jolokia.NewJolokia(pod.Status.PodIP, "8161", "/console/jolokia", user, pass)
					artemis = continuity.NewArtemisContinuity("amq-broker", jolokiaClient)
				}
			}

			artemisArray = append(artemisArray, artemis)
		}
	}

	return artemisArray
}

func retrieveBrokerCredentials(namespace string, client client.Client) (string, string, error) {
	secretName := secrets.CredentialsNameBuilder.Name()
	stringData := map[string]string{}
	secretNSName := types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}
	podSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    selectors.LabelBuilder.Labels(),
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: stringData,
	}

	err := resources.Retrieve(secretNSName, client, podSecret)
	user := string(podSecret.Data["AMQ_USER"])
	pass := string(podSecret.Data["AMQ_PASSWORD"])

	return user, pass, err
}
