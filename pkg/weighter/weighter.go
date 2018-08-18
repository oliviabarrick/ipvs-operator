package weighter

import (
	"github.com/justinbarrick/ipvs-operator/pkg/ipvs"
	"github.com/justinbarrick/ipvs-operator/pkg/utils"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

type ServiceWeighter struct {
	Balancer ipvs.Balancer
}

func NewIPVSServiceWeighter() (*ServiceWeighter, error) {
	ipvs, err := ipvs.NewIPVS()
	if err != nil {
		return nil, err
	}

	return &ServiceWeighter{
		Balancer: ipvs,
	}, nil
}

func (s *ServiceWeighter) SetForPod(pod corev1.Pod, service corev1.Service, weight int) error {
	if pod.Status.PodIP == "" {
		logrus.Infof("Waiting for pod %s to have IP address assigned.", pod.ObjectMeta.Name)
		return nil
	}

	for _, port := range service.Spec.Ports {
		targetPort := utils.GetTargetPort(port, pod)

		err := s.Balancer.SetWeight(service.Spec.ClusterIP, int(port.Port), pod.Status.PodIP,
			targetPort, string(port.Protocol), weight)
		if err != nil {
			logrus.Errorf("failed to set weight: %v", err)
			return err
		}
	}

	return nil
}
