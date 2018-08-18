package utils

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func newService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{},
	}
}

func newPod(containers []corev1.Container) corev1.Pod {
	return corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}
}

func TestMergeMaps(t *testing.T) {
	assert.Equal(t, MergeMaps(map[string]string{
		"hello": "world",
		"bye":   "moon",
	}, map[string]string{
		"hello": "mars",
		"hi":    "sun",
	}), map[string]string{
		"hello": "mars",
		"hi":    "sun",
		"bye":   "moon",
	})
}

func TestGetTargetPort(t *testing.T) {
	assert.Equal(t, GetTargetPort(corev1.ServicePort{
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromString("www"),
	}, newPod([]corev1.Container{
		corev1.Container{
			Ports: []corev1.ContainerPort{},
		},
		corev1.Container{
			Ports: []corev1.ContainerPort{
				corev1.ContainerPort{
					ContainerPort: int32(12346),
					Protocol:      corev1.ProtocolTCP,
				},
				corev1.ContainerPort{
					Name:          "www",
					ContainerPort: int32(12345),
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	})), 12345)

	assert.Equal(t, GetTargetPort(corev1.ServicePort{
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromString("www"),
	}, newPod([]corev1.Container{
		corev1.Container{
			Ports: []corev1.ContainerPort{},
		},
	})), 0)

	assert.Equal(t, GetTargetPort(corev1.ServicePort{
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(12345),
	}, newPod([]corev1.Container{})), 12345)

	assert.Equal(t, GetTargetPort(corev1.ServicePort{
		Protocol: corev1.ProtocolTCP,
		Port:     12345,
	}, newPod([]corev1.Container{})), 12345)

}

func TestHashObject(t *testing.T) {
	cr := newService()
	assert.Equal(t, HashObject(cr), "0517fe830bbcc4dd4283bbc5e9d01a99c263e982")
	cr.ObjectMeta.Name = "hello"
	assert.Equal(t, HashObject(cr), "5a58fb68ca6fd79618cdfb91af2945714d93fbb4")
}

func TestObjectHash(t *testing.T) {
	cr := newService()
	assert.Equal(t, GetObjectHash(cr), "")
	SetObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "0517fe830bbcc4dd4283bbc5e9d01a99c263e982")
	ClearObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "")
}

func TestClearObjectHashDoesNothingIfNoHashSet(t *testing.T) {
	cr := newService()
	ClearObjectHash(cr)
}
