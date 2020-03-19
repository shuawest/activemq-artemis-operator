package activemqartemis

import (
	//"context"
	"fmt"

	"github.com/rh-messaging/activemq-artemis-operator/pkg/controller/activemqartemisscaledown"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/ingresses"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/routes"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/secrets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	brokerv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/environments"
	svc "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/services"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/volumes"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/selectors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strconv"
	"strings"
)

const (
	statefulSetNotUpdated           = 0
	statefulSetSizeUpdated          = 1 << 0
	statefulSetClusterConfigUpdated = 1 << 1
	statefulSetImageUpdated         = 1 << 2
	statefulSetPersistentUpdated    = 1 << 3
	statefulSetAioUpdated           = 1 << 4
	statefulSetCommonConfigUpdated  = 1 << 5
	statefulSetRequireLoginUpdated  = 1 << 6
	//statefulSetRoleUpdated          = 1 << 7
	statefulSetAcceptorsUpdated  = 1 << 8
	statefulSetConnectorsUpdated = 1 << 9
	statefulSetConsoleUpdated    = 1 << 10
)

var defaultMessageMigration bool = true

type ActiveMQArtemisReconciler struct {
	statefulSetUpdates uint32
}

type ActiveMQArtemisIReconciler interface {
	Process(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet, firstTime bool) uint32
	ProcessContinuityPlugin(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet)
	ProcessDeploymentPlan(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet, firstTime bool) uint32
	ProcessAcceptors(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet)
	ProcessConnectors(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet)
	ProcessConsole(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet)
}

func (reconciler *ActiveMQArtemisReconciler) Process(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet, firstTime bool) uint32 {

	// TODO: Remove singular admin level user and password in favour of at least guest and admin access
	secretName := secrets.CredentialsNameBuilder.Name()

	envVarName := "AMQ_USER"
	adminUser := customResource.Spec.AdminUser
	if "" == adminUser {
		adminUser = environments.Defaults.AMQ_USER
	}
	statefulSetUpdates := sourceEnvVarFromSecret(customResource, currentStatefulSet, adminUser, envVarName, secretName, client, scheme)

	envVarName = "AMQ_PASSWORD"
	adminPassword := customResource.Spec.AdminPassword
	if "" == adminPassword {
		adminPassword = environments.Defaults.AMQ_PASSWORD
	}
	statefulSetUpdates |= sourceEnvVarFromSecret(customResource, currentStatefulSet, adminPassword, envVarName, secretName, client, scheme)

	statefulSetUpdates |= reconciler.ProcessContinuityPlugin(customResource, client, scheme, currentStatefulSet)

	statefulSetUpdates |= reconciler.ProcessDeploymentPlan(customResource, client, scheme, currentStatefulSet, firstTime)
	statefulSetUpdates |= reconciler.ProcessAcceptors(customResource, client, scheme, currentStatefulSet)
	statefulSetUpdates |= reconciler.ProcessConnectors(customResource, client, scheme, currentStatefulSet)
	statefulSetUpdates |= reconciler.ProcessConsole(customResource, client, scheme, currentStatefulSet)

	return statefulSetUpdates
}

func (reconciler *ActiveMQArtemisReconciler) ProcessContinuityPlugin(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet) uint32 {

	if customResource.Spec.EnableContinuity {
		secretName := secrets.NettyNameBuilder.Name()

		envVarName := "BROKER_ID_CACHE_SIZE"
		brokerIdCacheSize := customResource.Spec.BrokerIdCacheSize
		reconciler.statefulSetUpdates |= sourceEnvVarFromSecret(customResource, currentStatefulSet, strconv.Itoa(brokerIdCacheSize), envVarName, secretName, client, scheme)

		envVarName = "LOCAL_CONTINUITY_USER"
		reconciler.statefulSetUpdates |= sourceEnvVarFromSecret(customResource, currentStatefulSet, customResource.Spec.LocalContinuityUser, envVarName, secretName, client, scheme)
		envVarName = "LOCAL_CONTINUITY_PASS"
		reconciler.statefulSetUpdates |= sourceEnvVarFromSecret(customResource, currentStatefulSet, customResource.Spec.LocalContinuityPass, envVarName, secretName, client, scheme)

		// get list of acceptors that aren't for the cluster, scale controller, or continuity
		var servingAcceptors string = ""
		for _, acceptor := range customResource.Spec.Acceptors {
			if acceptor.Name != "continuity-external" {
				if servingAcceptors != "" {
					servingAcceptors = servingAcceptors + ";"
				}
				servingAcceptors = servingAcceptors + acceptor.Name
			}
		}

		// add continuity plugin config
		continuityPlugin := ""
		continuityPlugin = continuityPlugin + "<broker-plugins>\n"
		continuityPlugin = continuityPlugin + "   <broker-plugin class-name=\"org.apache.activemq.continuity.plugins.ContinuityPlugin\">\n"
		continuityPlugin = continuityPlugin + "      <property key=\"site-id\" value=\"" + customResource.Spec.SiteId + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"local-username\" value=\"" + customResource.Spec.LocalContinuityUser + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"local-password\" value=\"" + customResource.Spec.LocalContinuityPass + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"remote-username\" value=\"" + customResource.Spec.RemoteContinuityUser + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"remote-password\" value=\"" + customResource.Spec.RemoteContinuityPass + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"serving-acceptors\" value=\"" + servingAcceptors + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"local-connector-ref\" value=\"continuity-local\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"remote-connector-refs\" value=\"continuity-remote\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"active-on-start\" value=\"" + fmt.Sprintf("%v", customResource.Spec.ActiveOnStart) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"inflow-staging-delay-ms\" value=\"" + strconv.Itoa(customResource.Spec.InflowStagingDelay) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"bridge-interval-ms\" value=\"" + strconv.Itoa(customResource.Spec.BridgeInterval) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"bridge-interval-multiplier\" value=\"" + fmt.Sprintf("%f", customResource.Spec.BridgeIntervalMultiplier) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"outflow-exhausted-poll-duration-ms\" value=\"" + strconv.Itoa(customResource.Spec.OutflowExhaustedPollDuration) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"inflow-acks-consumed-poll-duration-ms\" value=\"" + strconv.Itoa(customResource.Spec.OutflowAcksConsumedPollDuration) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"activation-timeout-ms\" value=\"" + strconv.Itoa(customResource.Spec.ActivationTimeout) + "\" />\n"
		continuityPlugin = continuityPlugin + "      <property key=\"reorg-management-hierarchy\" value=\"" + fmt.Sprintf("%v", customResource.Spec.ReorgManagement) + "\" />\n"
		continuityPlugin = continuityPlugin + "   <\\/broker-plugin>\n"
		continuityPlugin = continuityPlugin + "<\\/broker-plugins>\n"

		envVarName = "CONTINUITY_PLUGIN"
		reconciler.statefulSetUpdates |= sourceEnvVarFromSecret(customResource, currentStatefulSet, continuityPlugin, envVarName, secretName, client, scheme)
	}

	return reconciler.statefulSetUpdates
}

func (reconciler *ActiveMQArtemisReconciler) ProcessDeploymentPlan(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet, firstTime bool) uint32 {

	deploymentPlan := &customResource.Spec.DeploymentPlan

	// Ensure the StatefulSet size is the same as the spec
	if *currentStatefulSet.Spec.Replicas != deploymentPlan.Size {
		currentStatefulSet.Spec.Replicas = &deploymentPlan.Size
		reconciler.statefulSetUpdates |= statefulSetSizeUpdated
	}

	if imageSyncCausedUpdateOn(deploymentPlan, currentStatefulSet) {
		reconciler.statefulSetUpdates |= statefulSetImageUpdated
	}

	if aioSyncCausedUpdateOn(deploymentPlan, currentStatefulSet) {
		reconciler.statefulSetUpdates |= statefulSetAioUpdated
	}

	if firstTime {
		if persistentSyncCausedUpdateOn(deploymentPlan, currentStatefulSet) {
			reconciler.statefulSetUpdates |= statefulSetPersistentUpdated
		}
	}

	if updatedEnvVar := environments.BoolSyncCausedUpdateOn(currentStatefulSet.Spec.Template.Spec.Containers, "AMQ_REQUIRE_LOGIN", deploymentPlan.RequireLogin); updatedEnvVar != nil {
		environments.Update(currentStatefulSet.Spec.Template.Spec.Containers, updatedEnvVar)
		reconciler.statefulSetUpdates |= statefulSetRequireLoginUpdated
	}

	syncMessageMigration(customResource, client, scheme)

	return reconciler.statefulSetUpdates
}

func (reconciler *ActiveMQArtemisReconciler) ProcessAcceptors(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet) uint32 {

	var retVal uint32 = statefulSetNotUpdated

	acceptorEntry := generateAcceptorsString(customResource, client)
	configureAcceptorsExposure(customResource, client, scheme)
	envVarName := "AMQ_ACCEPTORS"
	secretName := secrets.NettyNameBuilder.Name()
	retVal = sourceEnvVarFromSecret(customResource, currentStatefulSet, acceptorEntry, envVarName, secretName, client, scheme)

	return retVal
}

func (reconciler *ActiveMQArtemisReconciler) ProcessConnectors(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet) uint32 {

	var retVal uint32 = statefulSetNotUpdated

	connectorEntry := generateConnectorsString(customResource, client)
	configureConnectorsExposure(customResource, client, scheme)
	envVarName := "AMQ_CONNECTORS"
	secretName := secrets.NettyNameBuilder.Name()
	retVal = sourceEnvVarFromSecret(customResource, currentStatefulSet, connectorEntry, envVarName, secretName, client, scheme)

	return retVal
}

func (reconciler *ActiveMQArtemisReconciler) ProcessConsole(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, currentStatefulSet *appsv1.StatefulSet) uint32 {

	var retVal uint32 = statefulSetNotUpdated

	_, _ = configureConsoleExposure(customResource, client, scheme)
	if !customResource.Spec.Console.SSLEnabled {
		return retVal
	}

	sslFlags := ""
	envVarName := "AMQ_CONSOLE_ARGS"
	secretName := secrets.ConsoleNameBuilder.Name()
	if "" != customResource.Spec.Console.SSLSecret {
		secretName = customResource.Spec.Console.SSLSecret
	}
	sslFlags = generateConsoleSSLFlags(customResource, client, secretName)
	retVal = sourceEnvVarFromSecret(customResource, currentStatefulSet, sslFlags, envVarName, secretName, client, scheme)

	return retVal
}

func syncMessageMigration(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme) {

	var err error = nil
	var retrieveError error = nil

	namespacedName := types.NamespacedName{
		Name:      customResource.Name,
		Namespace: customResource.Namespace,
	}

	scaledown := &brokerv2alpha1.ActiveMQArtemisScaledown{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ActiveMQArtemisScaledown",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    selectors.LabelBuilder.Labels(),
			Name:      customResource.Name,
			Namespace: customResource.Namespace,
		},
		Spec: brokerv2alpha1.ActiveMQArtemisScaledownSpec{
			LocalOnly: true,
		},
		Status: brokerv2alpha1.ActiveMQArtemisScaledownStatus{},
	}

	if nil == customResource.Spec.DeploymentPlan.MessageMigration {
		customResource.Spec.DeploymentPlan.MessageMigration = &defaultMessageMigration
	}

	if *customResource.Spec.DeploymentPlan.MessageMigration {
		if err = resources.Retrieve(customResource, namespacedName, client, scaledown); err != nil {
			// err means not found so create
			if retrieveError = resources.Create(customResource, client, scheme, scaledown); retrieveError == nil {
			}
		}
	} else {
		if err = resources.Retrieve(customResource, namespacedName, client, scaledown); err == nil {
			close(activemqartemisscaledown.StopCh)
			// err means not found so delete
			if retrieveError = resources.Delete(customResource, client, scaledown); retrieveError == nil {
			}
		}
	}
}

func sourceEnvVarFromSecret(customResource *brokerv2alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet, acceptorEntry string, envVarName string, secretName string, client client.Client, scheme *runtime.Scheme) uint32 {

	var err error = nil
	var retVal uint32 = statefulSetNotUpdated

	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: currentStatefulSet.Namespace,
	}
	// Attempt to retrieve the secret
	stringDataMap := map[string]string{
		envVarName: acceptorEntry,
	}
	nettySecret := secrets.NewSecret(customResource, secretName, stringDataMap)
	if err = resources.Retrieve(customResource, namespacedName, client, nettySecret); err != nil {
		if errors.IsNotFound(err) {
			err = resources.Create(customResource, client, scheme, nettySecret)
		}
	} else { // err == nil so it already exists
		// Exists now
		// Check the contents against what we just got above
		elem, ok := nettySecret.Data[envVarName]
		if 0 != strings.Compare(string(elem), acceptorEntry) || !ok {
			nettySecret.Data[envVarName] = []byte(acceptorEntry)

			// These updates alone do not trigger a rolling update due to env var update as it's from a secret
			err = resources.Update(customResource, client, nettySecret)

			// Force the rolling update to occur
			environments.IncrementTriggeredRollCount(currentStatefulSet.Spec.Template.Spec.Containers)
			retVal = statefulSetAcceptorsUpdated
		}
	}

	acceptorsEnvVarSource := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
			Key:      envVarName,
			Optional: nil,
		},
	}
	acceptorsEnvVar := &corev1.EnvVar{
		Name:      envVarName,
		Value:     "",
		ValueFrom: acceptorsEnvVarSource,
	}
	if amqAcceptorsEnvVar := environments.Retrieve(currentStatefulSet.Spec.Template.Spec.Containers, envVarName); nil == amqAcceptorsEnvVar {
		environments.Create(currentStatefulSet.Spec.Template.Spec.Containers, acceptorsEnvVar)
	}

	return retVal
}

func generateAcceptorsString(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client) string {

	// TODO: Optimize for the single broker configuration
	ensureCOREOn61616Exists := true // as clustered is no longer an option but true by default

	acceptorEntry := ""
	defaultArgs := "tcpSendBufferSize=1048576;tcpReceiveBufferSize=1048576;useEpoll=true;amqpCredits=1000;amqpMinCredits=300"

	var portIncrement int32 = 10
	var currentPortIncrement int32 = 0
	var port61616InUse bool = false
	for _, acceptor := range customResource.Spec.Acceptors {
		if 0 == acceptor.Port {
			acceptor.Port = 61626 + currentPortIncrement
			currentPortIncrement += portIncrement
		}
		if "" == acceptor.Protocols ||
			"all" == strings.ToLower(acceptor.Protocols) {
			acceptor.Protocols = "AMQP,CORE,HORNETQ,MQTT,OPENWIRE,STOMP"
		}
		acceptorEntry = acceptorEntry + "<acceptor name=\"" + acceptor.Name + "\">"
		acceptorEntry = acceptorEntry + "tcp:" + "\\/\\/" + "ACCEPTOR_IP:"
		acceptorEntry = acceptorEntry + fmt.Sprintf("%d", acceptor.Port)
		acceptorEntry = acceptorEntry + "?protocols=" + strings.ToUpper(acceptor.Protocols)
		// TODO: Evaluate more dynamic messageMigration
		if 61616 == acceptor.Port {
			port61616InUse = true
		}
		if ensureCOREOn61616Exists &&
			(61616 == acceptor.Port) &&
			!strings.Contains(acceptor.Protocols, "CORE") {
			acceptorEntry = acceptorEntry + ",CORE"
		}
		if acceptor.SSLEnabled {
			secretName := customResource.Name + "-" + acceptor.Name + "-secret"
			if "" != acceptor.SSLSecret {
				secretName = acceptor.SSLSecret
			}
			acceptorEntry = acceptorEntry + ";" + generateAcceptorConnectorSSLArguments(customResource, client, secretName)
			sslOptionalArguments := generateAcceptorSSLOptionalArguments(acceptor)
			if "" != sslOptionalArguments {
				acceptorEntry = acceptorEntry + ";" + sslOptionalArguments
			}
		}
		if "" != acceptor.AnycastPrefix {
			safeAnycastPrefix := strings.Replace(acceptor.AnycastPrefix, "/", "\\/", -1)
			acceptorEntry = acceptorEntry + ";" + "anycastPrefix=" + safeAnycastPrefix
		}
		if "" != acceptor.MulticastPrefix {
			safeMulticastPrefix := strings.Replace(acceptor.MulticastPrefix, "/", "\\/", -1)
			acceptorEntry = acceptorEntry + ";" + "multicastPrefix=" + safeMulticastPrefix
		}
		if acceptor.ConnectionsAllowed > 0 {
			acceptorEntry = acceptorEntry + ";" + "connectionsAllowed=" + fmt.Sprintf("%d", acceptor.ConnectionsAllowed)
		}
		acceptorEntry = acceptorEntry + ";" + defaultArgs
		// TODO: SSL
		acceptorEntry = acceptorEntry + "<\\/acceptor>"
	}
	// TODO: Evaluate more dynamic messageMigration
	if ensureCOREOn61616Exists && !port61616InUse {
		acceptorEntry = acceptorEntry + "<acceptor name=\"" + "scaleDown" + "\">"
		acceptorEntry = acceptorEntry + "tcp:" + "\\/\\/" + "ACCEPTOR_IP:"
		acceptorEntry = acceptorEntry + fmt.Sprintf("%d", 61616)
		acceptorEntry = acceptorEntry + "?protocols=" + "CORE"
		acceptorEntry = acceptorEntry + ";" + defaultArgs
		// TODO: SSL
		acceptorEntry = acceptorEntry + "<\\/acceptor>"
	}

	if customResource.Spec.EnableContinuity {
		log.Info("Adding continuity-interal acceptor")
		acceptorEntry = acceptorEntry + "<acceptor name=\"continuity-internal\">vm:\\/\\/1<\\/acceptor>"

		// TODO: verify continuity-external acceptor exists
	}

	return acceptorEntry
}

func generateConnectorsString(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client) string {

	connectorEntry := ""
	connectors := customResource.Spec.Connectors
	for _, connector := range connectors {
		if connector.Type == "" {
			connector.Type = "tcp"
		}
		connectorEntry = connectorEntry + "<connector name=\"" + connector.Name + "\">"
		connectorEntry = connectorEntry + strings.ToLower(connector.Type) + ":\\/\\/" + strings.ToLower(connector.Host) + ":"
		connectorEntry = connectorEntry + fmt.Sprintf("%d", connector.Port)

		if connector.SSLEnabled {
			secretName := customResource.Name + "-" + connector.Name + "-secret"
			if "" != connector.SSLSecret {
				secretName = connector.SSLSecret
			}
			connectorEntry = connectorEntry + ";" + generateAcceptorConnectorSSLArguments(customResource, client, secretName)
			sslOptionalArguments := generateConnectorSSLOptionalArguments(connector)
			if "" != sslOptionalArguments {
				connectorEntry = connectorEntry + ";" + sslOptionalArguments
			}
		}
		connectorEntry = connectorEntry + "<\\/connector>"
	}

	if customResource.Spec.EnableContinuity {
		log.Info("Adding continuity-local connector")
		connectorEntry = connectorEntry + "<connector name=\"continuity-local\">vm:\\/\\/1<\\/connector>"

		// TODO: verify continuity-remote connector exists
	}

	return connectorEntry
}

func configureAcceptorsExposure(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme) (bool, error) {

	var i int32 = 0
	var err error = nil
	ordinalString := ""
	causedUpdate := false

	originalLabels := selectors.LabelBuilder.Labels()
	serviceRoutelabels := map[string]string{}
	for k, v := range originalLabels {
		serviceRoutelabels[k] = v
	}
	for ; i < customResource.Spec.DeploymentPlan.Size; i++ {
		ordinalString = strconv.Itoa(int(i))
		serviceRoutelabels["statefulset.kubernetes.io/pod-name"] = statefulsets.NameBuilder.Name() + "-" + ordinalString

		for _, acceptor := range customResource.Spec.Acceptors {
			serviceDefinition := svc.NewServiceDefinitionForCR(customResource, acceptor.Name+"-"+ordinalString, acceptor.Port, serviceRoutelabels)
			serviceNamespacedName := types.NamespacedName{
				Name:      serviceDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if acceptor.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
			}
			targetPortName := acceptor.Name + "-" + ordinalString
			targetServiceName := customResource.Name + "-" + targetPortName + "-svc"
			routeDefinition := routes.NewRouteDefinitionForCR(customResource, serviceRoutelabels, targetServiceName, targetPortName, acceptor.SSLEnabled)
			routeNamespacedName := types.NamespacedName{
				Name:      routeDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if acceptor.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			}
		}
	}

	return causedUpdate, err
}

func configureConnectorsExposure(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme) (bool, error) {

	var i int32 = 0
	var err error = nil
	ordinalString := ""
	causedUpdate := false

	originalLabels := selectors.LabelBuilder.Labels()
	serviceRoutelabels := map[string]string{}
	for k, v := range originalLabels {
		serviceRoutelabels[k] = v
	}
	for ; i < customResource.Spec.DeploymentPlan.Size; i++ {
		ordinalString = strconv.Itoa(int(i))
		serviceRoutelabels["statefulset.kubernetes.io/pod-name"] = statefulsets.NameBuilder.Name() + "-" + ordinalString

		for _, connector := range customResource.Spec.Connectors {
			serviceDefinition := svc.NewServiceDefinitionForCR(customResource, connector.Name+"-"+ordinalString, connector.Port, serviceRoutelabels)
			serviceNamespacedName := types.NamespacedName{
				Name:      serviceDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if connector.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
			}
			targetPortName := connector.Name + "-" + ordinalString
			targetServiceName := customResource.Name + "-" + targetPortName + "-svc"
			routeDefinition := routes.NewRouteDefinitionForCR(customResource, serviceRoutelabels, targetServiceName, targetPortName, connector.SSLEnabled)
			routeNamespacedName := types.NamespacedName{
				Name:      routeDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if connector.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			}
		}
	}

	return causedUpdate, err
}

func configureConsoleExposure(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme) (bool, error) {

	var i int32 = 0
	var err error = nil
	ordinalString := ""
	causedUpdate := false
	console := customResource.Spec.Console

	originalLabels := selectors.LabelBuilder.Labels()
	serviceRoutelabels := map[string]string{}
	for k, v := range originalLabels {
		serviceRoutelabels[k] = v
	}
	for ; i < customResource.Spec.DeploymentPlan.Size; i++ {
		ordinalString = strconv.Itoa(int(i))
		serviceRoutelabels["statefulset.kubernetes.io/pod-name"] = statefulsets.NameBuilder.Name() + "-" + ordinalString

		portNumber := int32(8161)
		targetPortName := "wconsj" + "-" + ordinalString
		targetServiceName := customResource.Name + "-" + targetPortName + "-svc"
		serviceDefinition := svc.NewServiceDefinitionForCR(customResource, targetPortName, portNumber, serviceRoutelabels)

		serviceNamespacedName := types.NamespacedName{
			Name:      serviceDefinition.Name,
			Namespace: customResource.Namespace,
		}
		if console.Expose {
			causedUpdate, err = resources.Enable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
		} else {
			causedUpdate, err = resources.Disable(customResource, client, scheme, serviceNamespacedName, serviceDefinition)
		}
		var err error = nil
		isOpenshift := false

		if isOpenshift, err = environments.DetectOpenshift(); err != nil {
			log.Error(err, "Failed to get env, will try kubernetes")
		}
		if isOpenshift {
			log.Info("Environment is OpenShift")
			log.Info("Checking routeDefinition for " + targetPortName)
			routeDefinition := routes.NewRouteDefinitionForCR(customResource, serviceRoutelabels, targetServiceName, targetPortName, console.SSLEnabled)
			routeNamespacedName := types.NamespacedName{
				Name:      routeDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if console.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, routeNamespacedName, routeDefinition)
			}
		} else {
			log.Info("Environment is not OpenShift, creating ingress")
			ingressDefinition := ingresses.NewIngressForCR(customResource, "wconsj")
			ingressNamespacedName := types.NamespacedName{
				Name:      ingressDefinition.Name,
				Namespace: customResource.Namespace,
			}
			if console.Expose {
				causedUpdate, err = resources.Enable(customResource, client, scheme, ingressNamespacedName, ingressDefinition)
			} else {
				causedUpdate, err = resources.Disable(customResource, client, scheme, ingressNamespacedName, ingressDefinition)
			}
		}
	}

	return causedUpdate, err
}

func generateConsoleSSLFlags(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, secretName string) string {

	sslFlags := ""
	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: customResource.Namespace,
	}
	stringDataMap := map[string]string{}
	userPasswordSecret := secrets.NewSecret(customResource, secretName, stringDataMap)

	keyStorePassword := "password"
	keyStorePath := "/etc/" + secretName + "-volume/broker.ks"
	trustStorePassword := "password"
	trustStorePath := "/etc/" + secretName + "-volume/client.ts"
	if err := resources.Retrieve(customResource, namespacedName, client, userPasswordSecret); err == nil {
		if "" != string(userPasswordSecret.Data["keyStorePassword"]) {
			keyStorePassword = string(userPasswordSecret.Data["keyStorePassword"])
		}
		if "" != string(userPasswordSecret.Data["keyStorePath"]) {
			keyStorePath = string(userPasswordSecret.Data["keyStorePath"])
		}
		if "" != string(userPasswordSecret.Data["trustStorePassword"]) {
			trustStorePassword = string(userPasswordSecret.Data["trustStorePassword"])
		}
		if "" != string(userPasswordSecret.Data["trustStorePath"]) {
			trustStorePath = string(userPasswordSecret.Data["trustStorePath"])
		}
	}

	sslFlags = sslFlags + " " + "--ssl-key" + " " + keyStorePath
	sslFlags = sslFlags + " " + "--ssl-key-password" + " " + keyStorePassword
	sslFlags = sslFlags + " " + "--ssl-trust" + " " + trustStorePath
	sslFlags = sslFlags + " " + "--ssl-trust-password" + " " + trustStorePassword
	if customResource.Spec.Console.UseClientAuth {
		sslFlags = sslFlags + " " + "--use-client-auth"
	}

	return sslFlags
}

func generateAcceptorConnectorSSLArguments(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, secretName string) string {

	sslArguments := "sslEnabled=true"
	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: customResource.Namespace,
	}
	stringDataMap := map[string]string{}
	userPasswordSecret := secrets.NewSecret(customResource, secretName, stringDataMap)

	keyStorePassword := "password"
	keyStorePath := "\\/etc\\/" + secretName + "-volume\\/broker.ks"
	trustStorePassword := "password"
	trustStorePath := "\\/etc\\/" + secretName + "-volume\\/client.ts"
	if err := resources.Retrieve(customResource, namespacedName, client, userPasswordSecret); err == nil {
		if "" != string(userPasswordSecret.Data["keyStorePassword"]) {
			keyStorePassword = string(userPasswordSecret.Data["keyStorePassword"])
		}
		if "" != string(userPasswordSecret.Data["keyStorePath"]) {
			keyStorePath = string(userPasswordSecret.Data["keyStorePath"])
		}
		if "" != string(userPasswordSecret.Data["trustStorePassword"]) {
			trustStorePassword = string(userPasswordSecret.Data["trustStorePassword"])
		}
		if "" != string(userPasswordSecret.Data["trustStorePath"]) {
			trustStorePath = string(userPasswordSecret.Data["trustStorePath"])
		}
	}
	sslArguments = sslArguments + ";" + "keyStorePath=" + keyStorePath
	sslArguments = sslArguments + ";" + "keyStorePassword=" + keyStorePassword
	sslArguments = sslArguments + ";" + "trustStorePath=" + trustStorePath
	sslArguments = sslArguments + ";" + "trustStorePassword=" + trustStorePassword

	return sslArguments
}

func generateAcceptorSSLOptionalArguments(acceptor brokerv2alpha1.AcceptorType) string {

	sslOptionalArguments := ""

	if "" != acceptor.EnabledCipherSuites {
		sslOptionalArguments = sslOptionalArguments + "enabledCipherSuites=" + acceptor.EnabledCipherSuites
	}
	if "" != acceptor.EnabledProtocols {
		sslOptionalArguments = sslOptionalArguments + ";" + "enabledProtocols=" + acceptor.EnabledProtocols
	}
	if acceptor.NeedClientAuth {
		sslOptionalArguments = sslOptionalArguments + ";" + "needClientAuth=true"
	}
	if acceptor.WantClientAuth {
		sslOptionalArguments = sslOptionalArguments + ";" + "wantClientAuth=true"
	}
	if acceptor.VerifyHost {
		sslOptionalArguments = sslOptionalArguments + ";" + "verifyHost=true"
	}
	if "" != acceptor.SSLProvider {
		sslOptionalArguments = sslOptionalArguments + ";" + "sslProvider=" + acceptor.SSLProvider
	}
	if "" != acceptor.SNIHost {
		sslOptionalArguments = sslOptionalArguments + ";" + "sniHost=" + acceptor.SNIHost
	}

	return sslOptionalArguments
}

func generateConnectorSSLOptionalArguments(connector brokerv2alpha1.ConnectorType) string {

	sslOptionalArguments := ""

	if "" != connector.EnabledCipherSuites {
		sslOptionalArguments = sslOptionalArguments + "enabledCipherSuites=" + connector.EnabledCipherSuites
	}
	if "" != connector.EnabledProtocols {
		sslOptionalArguments = sslOptionalArguments + ";" + "enabledProtocols=" + connector.EnabledProtocols
	}
	if connector.NeedClientAuth {
		sslOptionalArguments = sslOptionalArguments + ";" + "needClientAuth=true"
	}
	if connector.WantClientAuth {
		sslOptionalArguments = sslOptionalArguments + ";" + "wantClientAuth=true"
	}
	if connector.VerifyHost {
		sslOptionalArguments = sslOptionalArguments + ";" + "verifyHost=true"
	}
	if "" != connector.SSLProvider {
		sslOptionalArguments = sslOptionalArguments + ";" + "sslProvider=" + connector.SSLProvider
	}
	if "" != connector.SNIHost {
		sslOptionalArguments = sslOptionalArguments + ";" + "sniHost=" + connector.SNIHost
	}

	return sslOptionalArguments
}

// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
func remove(s []corev1.EnvVar, i int) []corev1.EnvVar {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

func aioSyncCausedUpdateOn(deploymentPlan *brokerv2alpha1.DeploymentPlanType, currentStatefulSet *appsv1.StatefulSet) bool {

	foundAio := false
	foundNio := false
	var extraArgs string = ""
	extraArgsNeedsUpdate := false

	// Find the existing values
	for _, v := range currentStatefulSet.Spec.Template.Spec.Containers[0].Env {
		if v.Name == "AMQ_JOURNAL_TYPE" {
			if strings.Index(v.Value, "aio") > -1 {
				foundAio = true
			}
			if strings.Index(v.Value, "nio") > -1 {
				foundNio = true
			}
			extraArgs = v.Value
			break
		}
	}

	if "aio" == strings.ToLower(deploymentPlan.JournalType) && foundNio {
		extraArgs = strings.Replace(extraArgs, "nio", "aio", 1)
		extraArgsNeedsUpdate = true
	}

	if !("aio" == strings.ToLower(deploymentPlan.JournalType)) && foundAio {
		extraArgs = strings.Replace(extraArgs, "aio", "nio", 1)
		extraArgsNeedsUpdate = true
	}

	if !foundAio && !foundNio {
		extraArgs = "--" + strings.ToLower(deploymentPlan.JournalType)
		extraArgsNeedsUpdate = true
	}

	if extraArgsNeedsUpdate {
		newExtraArgsValue := corev1.EnvVar{
			"AMQ_JOURNAL_TYPE",
			extraArgs,
			nil,
		}
		environments.Update(currentStatefulSet.Spec.Template.Spec.Containers, &newExtraArgsValue)
	}

	return extraArgsNeedsUpdate
}

func persistentSyncCausedUpdateOn(deploymentPlan *brokerv2alpha1.DeploymentPlanType, currentStatefulSet *appsv1.StatefulSet) bool {

	foundDataDir := false
	foundDataDirLogging := false

	dataDirNeedsUpdate := false
	dataDirLoggingNeedsUpdate := false

	statefulSetUpdated := false

	// TODO: Remove yuck
	// ensure password and username are valid if can't via openapi validation?
	if deploymentPlan.PersistenceEnabled {

		envVarArray := []corev1.EnvVar{}
		// Find the existing values
		for _, v := range currentStatefulSet.Spec.Template.Spec.Containers[0].Env {
			if v.Name == "AMQ_DATA_DIR" {
				foundDataDir = true
				if v.Value != volumes.GLOBAL_DATA_PATH {
					dataDirNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_DATA_DIR_LOGGING" {
				foundDataDirLogging = true
				if v.Value != "true" {
					dataDirLoggingNeedsUpdate = true
				}
			}
		}

		if !foundDataDir || dataDirNeedsUpdate {
			newDataDirValue := corev1.EnvVar{
				"AMQ_DATA_DIR",
				volumes.GLOBAL_DATA_PATH,
				nil,
			}
			envVarArray = append(envVarArray, newDataDirValue)
			statefulSetUpdated = true
		}

		if !foundDataDirLogging || dataDirLoggingNeedsUpdate {
			newDataDirLoggingValue := corev1.EnvVar{
				"AMQ_DATA_DIR_LOGGING",
				"true",
				nil,
			}
			envVarArray = append(envVarArray, newDataDirLoggingValue)
			statefulSetUpdated = true
		}

		if statefulSetUpdated {
			envVarArrayLen := len(envVarArray)
			if envVarArrayLen > 0 {
				for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
					for j := len(currentStatefulSet.Spec.Template.Spec.Containers[i].Env) - 1; j >= 0; j-- {
						if ("AMQ_DATA_DIR" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name && dataDirNeedsUpdate) ||
							("AMQ_DATA_DIR_LOGGING" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name && dataDirLoggingNeedsUpdate) {
							currentStatefulSet.Spec.Template.Spec.Containers[i].Env = remove(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, j)
						}
					}
				}

				containerArrayLen := len(currentStatefulSet.Spec.Template.Spec.Containers)
				for i := 0; i < containerArrayLen; i++ {
					for j := 0; j < envVarArrayLen; j++ {
						currentStatefulSet.Spec.Template.Spec.Containers[i].Env = append(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, envVarArray[j])
					}
				}
			}
		}
	} else {

		for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
			for j := len(currentStatefulSet.Spec.Template.Spec.Containers[i].Env) - 1; j >= 0; j-- {
				if "AMQ_DATA_DIR" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_DATA_DIR_LOGGING" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name {
					currentStatefulSet.Spec.Template.Spec.Containers[i].Env = remove(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, j)
					statefulSetUpdated = true
				}
			}
		}
	}

	return statefulSetUpdated
}

func imageSyncCausedUpdateOn(deploymentPlan *brokerv2alpha1.DeploymentPlanType, currentStatefulSet *appsv1.StatefulSet) bool {

	// At implementation time only one container
	if strings.Compare(currentStatefulSet.Spec.Template.Spec.Containers[0].Image, deploymentPlan.Image) != 0 {
		containerArrayLen := len(currentStatefulSet.Spec.Template.Spec.Containers)
		for i := 0; i < containerArrayLen; i++ {
			currentStatefulSet.Spec.Template.Spec.Containers[i].Image = deploymentPlan.Image
		}
		return true
	}

	return false
}
