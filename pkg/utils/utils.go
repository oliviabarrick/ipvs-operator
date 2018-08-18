package utils

import (
	"fmt"
	"github.com/cnf/structhash"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Merge two maps. The keys in second will overwrite any matching keys in base.
func MergeMaps(base map[string]string, second map[string]string) map[string]string {
	merged := map[string]string{}

	for k, v := range base {
		merged[k] = v
	}

	for k, v := range second {
		merged[k] = v
	}

	return merged
}

// Return an integer target port from a ServicePort for a pod.
func GetTargetPort(port corev1.ServicePort, pod corev1.Pod) int {
	target := port.Port

	if port.TargetPort.Type == intstr.Int {
		if port.TargetPort.IntVal != 0 {
			target = port.TargetPort.IntVal
		}
	} else {
		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				if containerPort.Name == port.TargetPort.StrVal {
					target = containerPort.ContainerPort
				}
			}
		}
	}

	return int(target)
}

// Takes a Kubernetes object and returns the hash in its annotations as a string.
func GetObjectHash(obj runtime.Object) string {
	objectMeta, _ := meta.Accessor(obj)

	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	return annotations["weightedservice.codesink.net.hash"]
}

// Takes a Kubernetes object and adds an annotation with its hash.
func SetObjectHash(obj runtime.Object) {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		fmt.Println(err)
	}

	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["weightedservice.codesink.net.hash"] = HashObject(obj)
	objectMeta.SetAnnotations(annotations)
}

// Takes a Kubernetes object and removes the annotation with its hash.
func ClearObjectHash(obj runtime.Object) {
	objectMeta, _ := meta.Accessor(obj)

	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		return
	}

	delete(annotations, "weightedservice.codesink.net.hash")
	objectMeta.SetAnnotations(annotations)
}

// Return a SHA1 hash of a Kubernetes object
func HashObject(obj runtime.Object) string {
	copied := obj.DeepCopyObject()
	objectMeta, _ := meta.Accessor(copied)

	ownerReferences := objectMeta.GetOwnerReferences()
	if len(ownerReferences) > 0 {
		ownerReferences[0].UID = ""
	}
	objectMeta.SetOwnerReferences(ownerReferences)

	return fmt.Sprintf("%x", structhash.Sha1(copied, 1))
}
