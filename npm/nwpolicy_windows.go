// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"github.com/kalebmorris/azure-container-networking/log"
	"github.com/kalebmorris/azure-container-networking/npm/util"
	"github.com/kalebmorris/azure-container-networking/npm/vfpm"
	networkingv1 "k8s.io/api/networking/v1"
)

// AddNetworkPolicy handles adding network policy to vfp.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	var err error

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	log.Printf("NETWORK POLICY CREATING: %v", npObj)

	allNs := npMgr.nsMap[util.KubeAllNamespacesFlag]

	ports, err := vfpm.GetPorts()
	if err != nil {
		log.Errorf("Error: failed to retrieve ports.")
		return err
	}

	for _, portName := range ports {
		if err = allNs.tMgr.CreateTag(util.KubeSystemFlag, portName); err != nil {
			log.Errorf("Error: failed to initialize kube-system tag.")
			return err
		}
	}

	podTags, nsLists, rules := parsePolicy(npObj)

	tMgr := allNs.tMgr
	rMgr := allNs.rMgr

	for _, portName := range ports {
		for _, tag := range podTags {
			if err = tMgr.CreateTag(tag, portName); err != nil {
				log.Errorf("Error: failed to create tag %s-%s on port %s.", npNs, tag, portName)
				return err
			}
		}

		for _, nlTag := range nsLists {
			if err = tMgr.CreateNLTag(nlTag, portName); err != nil {
				log.Errorf("Error: failed to create NLTag %s-%s on port %s.", npNs, nlTag, portName)
				return err
			}
		}

		if err = npMgr.InitAllNsList(portName); err != nil {
			log.Errorf("Error: failed to initialize all-namespace NLTag on port %s.", portName)
			return err
		}

		for _, rule := range rules {
			if err = rMgr.Add(rule, portName); err != nil {
				log.Errorf("Error: failed to apply rule on port %s. Rule: %+v", portName, rule)
				return err
			}
		}
	}

	allNs.npMap[npName] = npObj

	ns, err := newNs(npNs)
	if err != nil {
		log.Errorf("Error: failed to create namespace %s", npNs)
	}
	npMgr.nsMap[npNs] = ns

	return nil
}

// DeleteNetworkPolicy handles deleting network policy from hcn.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	var err error

	npName := npObj.ObjectMeta.Name
	log.Printf("NETWORK POLICY DELETING: %v", npObj)

	allNs := npMgr.nsMap[util.KubeAllNamespacesFlag]

	ports, err := vfpm.GetPorts()
	if err != nil {
		log.Errorf("Error: failed to retrieve ports.")
		return err
	}

	_, _, rules := parsePolicy(npObj)

	rMgr := allNs.rMgr

	for _, portName := range ports {
		for _, rule := range rules {
			if err = rMgr.Delete(rule, portName); err != nil {
				log.Errorf("Error: failed to delete rule. Rule: %+v", rule)
				return err
			}
		}
	}

	delete(allNs.npMap, npName)

	return nil
}