package ipvs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mqliang/libipvs"
	"github.com/sirupsen/logrus"
)

type Balancer interface {
	SetScheduler(service string, servicePort int, protocol string, scheduler string) error
	SetWeight(service string, servicePort int, destination string, destinationPort int, protocol string, weight int) error
}

type IPVS struct {
	ipvs libipvs.IPVSHandle
}

func NewIPVS() (*IPVS, error) {
	ipvs, err := libipvs.NewIPVSHandle(libipvs.IPVSHandleParams{})
	if err != nil {
		return nil, err
	}

	return &IPVS{
		ipvs: ipvs,
	}, nil
}

func (i *IPVS) getService(service string, port int, protocol string) (*libipvs.Service, error) {
	svcs, err := i.ipvs.ListServices()
	if err != nil {
		return nil, err
	}

	for _, svc := range svcs {
		if svc.Address.String() != service {
			continue
		}

		if svc.Port != uint16(port) {
			fmt.Println("port mismatch")
			continue
		}

		if svc.Protocol.String() != strings.ToLower(protocol) {
			fmt.Println("protocol mismatch")
			continue
		}

		return svc, nil
	}

	return nil, errors.New("Could not find service with that address.")
}

func (i *IPVS) getDestination(svc *libipvs.Service, destination string, destinationPort int) (*libipvs.Destination, error) {
	dsts, err := i.ipvs.ListDestinations(svc)
	if err != nil {
		return nil, err
	}

	for _, dst := range dsts {
		if dst.Address.String() != destination {
			continue
		}

		if dst.Port != uint16(destinationPort) {
			continue
		}

		return dst, nil
	}

	return nil, errors.New("Could not find destination with that address.")
}

func (i *IPVS) SetScheduler(service string, servicePort int, protocol string, scheduler string) error {
	svc, err := i.getService(service, servicePort, protocol)
	if err != nil {
		return err
	}

	svc.SchedName = scheduler
	return i.ipvs.UpdateService(svc)
}

func (i *IPVS) SetWeight(service string, servicePort int, destination string, destinationPort int, protocol string, weight int) error {
	svc, err := i.getService(service, servicePort, protocol)
	if err != nil {
		return err
	}

	dst, err := i.getDestination(svc, destination, destinationPort)
	if err != nil {
		return err
	}

	if dst.Weight == uint32(weight) {
		return nil
	}

	logrus.Infof("Setting weight: Service IP: %s:%d, Pod IP: %s:%d, Protocol: %s, Weight: %d",
		service, servicePort, destination, destinationPort, protocol, weight)

	dst.Weight = uint32(weight)
	return i.ipvs.UpdateDestination(svc, dst)
}
