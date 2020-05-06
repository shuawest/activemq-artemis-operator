package v2alpha1activemqartemisaddress

import (
	"context"
	"strconv"
	"strings"
	"time"

	brokerv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	brokerv2alpha2 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2"
	mgmt "github.com/rh-messaging/activemq-artemis-operator/pkg/management"
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

var log = logf.Log.WithName("controller_v2alpha1activemqartemisaddress")
var namespacedNameToAddressName = make(map[types.NamespacedName]brokerv2alpha1.ActiveMQArtemisAddress)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ActiveMQArtemisAddress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileActiveMQArtemisAddress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("v2alpha1activemqartemisaddress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemisAddress
	err = c.Watch(&source.Kind{Type: &brokerv2alpha1.ActiveMQArtemisAddress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemisAddress
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv2alpha1.ActiveMQArtemisAddress{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileActiveMQArtemisAddress{}

// ReconcileActiveMQArtemisAddress reconciles a ActiveMQArtemisAddress object
type ReconcileActiveMQArtemisAddress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ActiveMQArtemisAddress object and makes changes based on the state read
// and what is in the ActiveMQArtemisAddress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileActiveMQArtemisAddress) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemisAddress")

	// Fetch the ActiveMQArtemisAddress instance
	instance := &brokerv2alpha1.ActiveMQArtemisAddress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		// Delete action
		addressInstance, lookupSucceeded := namespacedNameToAddressName[request.NamespacedName]
		if lookupSucceeded {
			err = deleteQueue(&addressInstance, request, r.client)
			delete(namespacedNameToAddressName, request.NamespacedName)
		}
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	} else {
		// Fetch the ActiveMQArtemis broker CR
		brokerCR := &brokerv2alpha2.ActiveMQArtemis{}

		brokerNSName := types.NamespacedName{
			Name:      instance.GetClusterName(),
			Namespace: request.NamespacedName.Namespace,
		}
		err = r.client.Get(context.TODO(), brokerNSName, brokerCR)

		err = createQueue(instance, request, r.client)
		if nil == err {
			// Verify address is registered
			queuePresent, err := verifyQueue(instance, request, r.client)
			if nil != err || queuePresent == false {
				reqLogger.Info("ActiveMQArtemisAddress '" + instance.Spec.QueueName + "' has not been created on all brokers, requeuing reconcile")
				return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
			} else {
				// success - don't requeue the request
				reqLogger.Info("ActiveMQArtemisAddress '" + instance.Spec.QueueName + "' created in all brokers")
				namespacedNameToAddressName[request.NamespacedName] = *instance
				return reconcile.Result{}, nil
			}
		}
	}

	return reconcile.Result{}, err
}

func createQueue(instance *brokerv2alpha1.ActiveMQArtemisAddress, request reconcile.Request, client client.Client) error {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Creating ActiveMQArtemisAddress")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for _, a := range artemisArray {
			if nil == a {
				continue
			}
			err := a.CreateQueue(instance.Spec.AddressName, instance.Spec.QueueName, instance.Spec.RoutingType)
			if nil != err {
				reqLogger.Info("Creating ActiveMQArtemisAddress error for " + instance.Spec.QueueName + ": " + err.Error())
				break
			} else {
				reqLogger.Info("Created ActiveMQArtemisAddress for " + instance.Spec.QueueName)
			}
		}
	}

	return err
}

func verifyQueue(instance *brokerv2alpha1.ActiveMQArtemisAddress, request reconcile.Request, client client.Client) (bool, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Verifying ActiveMQArtemisAddress")

	queuePresent := true
	var err error = nil

	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for _, a := range artemisArray {
			if nil == a {
				reqLogger.Info("Verify ActiveMQArtemisAddress: broker not available yet", "QueueName", instance.Spec.QueueName)
				queuePresent = false
				continue
			}

			bindingsData, err := a.ListBindingsForAddress(instance.Spec.AddressName)
			if err != nil || "" == bindingsData || !strings.Contains(bindingsData, "name="+instance.Spec.QueueName+",") {
				reqLogger.Info("Verify ActiveMQArtemisAddress: queue not created yet", "QueueName", instance.Spec.QueueName)
				queuePresent = false
				continue
			}
		}
	}

	return queuePresent, err
}

func deleteQueue(instance *brokerv2alpha1.ActiveMQArtemisAddress, request reconcile.Request, client client.Client) error {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Deleting ActiveMQArtemisAddress")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for _, a := range artemisArray {
			err := a.DeleteQueue(instance.Spec.QueueName)
			if nil != err {
				reqLogger.Error(err, "Deleting ActiveMQArtemisAddress error for "+instance.Spec.QueueName)
				break
			} else {
				reqLogger.Info("Deleted ActiveMQArtemisAddress for " + instance.Spec.QueueName)
				reqLogger.Info("Checking parent address for bindings " + instance.Spec.AddressName)
				bindingsData, err := a.ListBindingsForAddress(instance.Spec.AddressName)
				if nil == err {
					if "" == bindingsData {
						reqLogger.Info("No bindings found removing " + instance.Spec.AddressName)
						a.DeleteAddress(instance.Spec.AddressName)
					} else {
						reqLogger.Info("Bindings found, not removing " + instance.Spec.AddressName)
					}
				}
			}
		}
	}

	return err
}

func getPodBrokers(instance *brokerv2alpha1.ActiveMQArtemisAddress, request reconcile.Request, client client.Client) []*mgmt.Artemis {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Getting Pod Brokers")

	var artemisArray []*mgmt.Artemis = make([]*mgmt.Artemis, 1)
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
		artemisArray = make([]*mgmt.Artemis, 0, replicas)
		for i = 0; i < replicas; i++ {
			s := statefulset.Name + "-" + strconv.Itoa(i)
			podNamespacedName.Name = s

			var artemis *mgmt.Artemis = nil

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
					artemis = mgmt.NewArtemis("amq-broker", jolokiaClient)
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
