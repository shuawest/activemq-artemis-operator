package serviceports

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
)

func GetDefaultPorts() *[]corev1.ServicePort {

	ports := []corev1.ServicePort{
		{
			Name:       "console-jolokia",
			Protocol:   "TCP",
			Port:       8161,
			TargetPort: intstr.FromInt(int(8161)),
		},
		{
			Name:       "all",
			Protocol:   "TCP",
			Port:       61616,
			TargetPort: intstr.FromInt(int(61616)),
		},
	}

	return &ports
}

func setSSLPorts() []corev1.ServicePort {

	ports := []corev1.ServicePort{

		{
			Name:       "amqp-ssl",
			Protocol:   "TCP",
			Port:       5671,
			TargetPort: intstr.FromInt(int(5671)),
		},
		{
			Name:       "mqtt-ssl",
			Protocol:   "TCP",
			Port:       8883,
			TargetPort: intstr.FromInt(int(8883)),
		},
		{
			Name:       "stomp-ssl",
			Protocol:   "TCP",
			Port:       61612,
			TargetPort: intstr.FromInt(int(61612)),
		},
	}

	return ports
}

func setBasicPorts() []corev1.ServicePort {

	ports := []corev1.ServicePort{
		{
			Name:       "mqtt",
			Protocol:   "TCP",
			Port:       1883,
			TargetPort: intstr.FromInt(int(1883)),
		},
		{
			Name:       "amqp",
			Protocol:   "TCP",
			Port:       5672,
			TargetPort: intstr.FromInt(int(5672)),
		},
		{
			Name:       "console-jolokia",
			Protocol:   "TCP",
			Port:       8161,
			TargetPort: intstr.FromInt(int(8161)),
		},
		{
			Name:       "stomp",
			Protocol:   "TCP",
			Port:       61613,
			TargetPort: intstr.FromInt(int(61613)),
		},
		{
			Name:       "all",
			Protocol:   "TCP",
			Port:       61616,
			TargetPort: intstr.FromInt(int(61616)),
		},
	}

	return ports
}
