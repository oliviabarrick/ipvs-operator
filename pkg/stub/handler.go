package stub

import (
	"context"

	"github.com/justinbarrick/ipvs-operator/pkg/apis/codesink/v1alpha1"
	"github.com/justinbarrick/ipvs-operator/pkg/utils"
	"github.com/justinbarrick/ipvs-operator/pkg/weighter"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// newService creates a Service for a WeightedService
func newService(cr *v1alpha1.WeightedService) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.ObjectMeta.Name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "WeightedService",
				}),
			},
		},
		Spec: *cr.Spec.ServiceSpec,
	}
}

// List pods in namespace with optional ListOptions
func listPods(namespace string, opts ...sdk.ListOption) ([]corev1.Pod, error) {
	pods := corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
	err := sdk.List(namespace, &pods, opts...)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

type Handler struct {
	weighter *weighter.ServiceWeighter
}

func NewHandler() sdk.Handler {
	weighter, err := weighter.NewIPVSServiceWeighter()
	if err != nil {
		logrus.Fatalf("failed to open IPVS interface: %v", err)
	}

	return &Handler{
		weighter: weighter,
	}
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.WeightedService:
		return h.syncWeightedService(o)
	case *corev1.Service:
		return h.syncService(o)
	}
	return nil
}

// Create the services needed by a WeightedService
func (h *Handler) syncWeightedService(service *v1alpha1.WeightedService) error {
	existingService := newService(service)

	desiredService := newService(service)
	utils.SetObjectHash(desiredService)

	err := sdk.Get(existingService)
	if err == nil {
		if utils.GetObjectHash(existingService) == utils.GetObjectHash(desiredService) {
			return nil
		}

		desiredService.Spec.ClusterIP = existingService.Spec.ClusterIP

		existingMeta, _ := meta.Accessor(existingService)
		desiredMeta, _ := meta.Accessor(desiredService)

		desiredMeta.SetResourceVersion(existingMeta.GetResourceVersion())

		logrus.Infof("Updating service %s", desiredService.ObjectMeta.Name)
		return sdk.Update(desiredService)
	} else {
		logrus.Infof("Creating service %s", desiredService.ObjectMeta.Name)
		return sdk.Create(desiredService)
	}
}

// If a Service is owned by a WeightedService, set the load balancer weights
// on the nodes for each pod as appropriate.
func (h *Handler) syncService(service *corev1.Service) error {
	if service.Spec.ClusterIP == "" {
		return nil
	}

	if len(service.ObjectMeta.OwnerReferences) == 0 {
		return nil
	}

	if service.ObjectMeta.OwnerReferences[0].Kind != "WeightedService" {
		return nil
	}

	owner := v1alpha1.WeightedService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "WeightedService",
			APIVersion: "codesink.net/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ObjectMeta.OwnerReferences[0].Name,
			Namespace: service.ObjectMeta.Namespace,
		},
	}
	err := sdk.Get(&owner)
	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Errorf("WeightedService service does not exist, cleaning up.")
			return nil
		}

		logrus.Errorf("failed to get WeightedService: %v", err)
		return err
	}

	scheduler := owner.Spec.Scheduler
	if scheduler == "" {
		scheduler = "wrr"
	}
	h.weighter.SetScheduler(*service, scheduler)

	for _, weight := range owner.Spec.Weights {
		pods, err := listPods(service.ObjectMeta.Namespace, sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: utils.MergeMaps(service.Spec.Selector, weight.Selector),
			}),
		}))
		if err != nil {
			logrus.Errorf("listing pods: %v", err)
			return err
		}

		for _, pod := range pods {
			h.weighter.SetForPod(pod, *service, weight.Weight)
		}
	}

	return nil
}
