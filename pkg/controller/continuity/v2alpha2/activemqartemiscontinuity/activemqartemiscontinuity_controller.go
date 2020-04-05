package activemqartemiscontinuity

import (
	"context"
	"strconv"

	continuityv2alpha2 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2"
	mgmt "github.com/rh-messaging/activemq-artemis-operator/pkg/continuity"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
var namespacedNameToContinuityName = make(map[types.NamespacedName]continuityv2alpha2.ActiveMQArtemisContinuity)

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
	reqLogger.Info("Reconciling ActiveMQArtemisContinuity 002")

	// Fetch the ActiveMQArtemisContinuity instance
	instance := &continuityv2alpha2.ActiveMQArtemisContinuity{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {

		if errors.IsNotFound(err) {
			reqLogger.Info("Reconciling ActiveMQArtemisContinuity 002 - IsNotFound no requeue")

			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		reqLogger.Info("Reconciling ActiveMQArtemisContinuity 002 - error do requeue")

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Reconciling ActiveMQArtemisContinuity 002 - success")

		err = createContinuity(instance, request, r.client)
		if nil == err {
			namespacedNameToContinuityName[request.NamespacedName] = *instance
		}

		if err != nil {
			reqLogger.Info("Reconciling ActiveMQArtemisContinuity 002 - err calling createContinuity " + err.Error())
		}
	}

	return reconcile.Result{}, nil
}

func createContinuity(instance *continuityv2alpha2.ActiveMQArtemisContinuity, request reconcile.Request, client client.Client) error {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Creating ActiveMQArtemisContinuity createContinuity")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)

	reqLogger.Info("Pod brokers: " + strconv.Itoa(len(artemisArray)))

	if nil != artemisArray {
		for _, a := range artemisArray {
			if nil == a {
				reqLogger.Info("Creating ActiveMQArtemisContinuity artemisArray had a nil!")
				continue
			}
			_, errConfigure := a.Configure(instance.Spec.SiteId, instance.Spec.ActiveOnStart, instance.Spec.ServingAcceptors, "continuity-local", instance.Spec.RemoteConnectorRefs, instance.Spec.ReorgManagement)
			_, errSecrets := a.SetSecrets(instance.Spec.LocalContinuityUser, instance.Spec.LocalContinuityPass, instance.Spec.RemoteContinuityUser, instance.Spec.RemoteContinuityPass)
			_, errTune := a.Tune(instance.Spec.ActivationTimeout, instance.Spec.InflowStagingDelay, instance.Spec.BridgeInterval, instance.Spec.BridgeIntervalMultiplier, instance.Spec.PollDuration)
			_, errBoot := a.Boot()
			if nil != errConfigure || nil != errSecrets || nil != errTune || nil != errBoot {
				reqLogger.Info("Creating ActiveMQArtemisContinuity error for " + instance.Spec.SiteId)
				break
			} else {
				reqLogger.Info("Created ActiveMQArtemisContinuity for " + instance.Spec.SiteId)
			}
		}
	}

	return err
}

func getPodBrokers(instance *continuityv2alpha2.ActiveMQArtemisContinuity, request reconcile.Request, client client.Client) []*mgmt.ArtemisContinuity {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Getting Pod Brokers")

	var artemisArray []*mgmt.ArtemisContinuity = nil
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

		// For each of the replicas
		var i int = 0
		var replicas int = int(*statefulset.Spec.Replicas)
		artemisArray = make([]*mgmt.ArtemisContinuity, 0, replicas)
		for i = 0; i < replicas; i++ {
			s := statefulset.Name + "-" + strconv.Itoa(i)
			podNamespacedName.Name = s
			if err = client.Get(context.TODO(), podNamespacedName, pod); err != nil {
				if errors.IsNotFound(err) {
					reqLogger.Error(err, "Pod IsNotFound", "Namespace", request.Namespace, "Name", request.Name)
				} else {
					reqLogger.Error(err, "Pod lookup error", "Namespace", request.Namespace, "Name", request.Name)
				}
			} else {
				reqLogger.Info("Pod found", "Namespace", request.Namespace, "Name", request.Name)
				artemis := mgmt.NewArtemisContinuity(pod.Status.PodIP, "8161", "amq-broker")
				artemisArray = append(artemisArray, artemis)
			}
		}
	}

	return artemisArray
}
