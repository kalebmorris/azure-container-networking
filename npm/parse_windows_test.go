// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kalebmorris/azure-container-networking/npm/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetTargetTags(t *testing.T) {
	netPol := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "test-nwpolicy",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "frontend",
					"purpose": "portal",
				},
			},
		},
	}

	reconstructed := make(map[string]string)
	targetTags := getTargetTags(netPol)
	for _, tag := range targetTags {
		idx := strings.Index(tag, util.KubeAllNamespacesFlag)
		if idx == -1 {
			continue
		}
		tag = tag[idx+len(util.KubeAllNamespacesFlag)+1:]
		idx = strings.Index(tag, ":")
		if idx == -1 {
			continue
		}
		key := tag[:idx]
		val := tag[idx+1:]
		reconstructed[key] = val
	}

	if !reflect.DeepEqual(netPol.Spec.PodSelector.MatchLabels, reconstructed) {
		t.Errorf("TestGetTargetTags failed")
	}
}

func TestGetPolicyTypes(t *testing.T) {
	bothPolTypes := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "test-nwpolicy",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	ingressExists, egressExists := getPolicyTypes(bothPolTypes)
	if !ingressExists || !egressExists {
		t.Errorf("TestGetPolicyTypes failed")
	}

	neitherPolTypes := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "test-nwpolicy",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{},
		},
	}

	ingressExists, egressExists = getPolicyTypes(neitherPolTypes)
	if ingressExists || egressExists {
		t.Errorf("TestGetPolicyTypes failed")
	}
}

func TestIpToInt(t *testing.T) {
	ip, err := ipToInt("0.1.2.3")
	if err != nil || ip != uint32(66051) {
		t.Errorf("TestIpToInt failed @ ipToInt")
	}

	ip, err = ipToInt("3.2.1.0")
	if err != nil || ip != uint32(50462976) {
		t.Errorf("TestIpToInt failed @ ipToInt")
	}
}

func TestGetRanges(t *testing.T) {
	ipblock := &networkingv1.IPBlock{
		CIDR: "10.240.6.6/16",
		Except: []string{
			"10.240.10.2/24",
			"10.240.11.4/24",
			"10.240.221.0/22",
			"10.235.0.0/30",
		},
	}

	starts, ends := getRanges(ipblock)

	startsIPs := []string{
		"10.240.0.0",
		"10.240.12.0",
		"10.240.224.0",
	}
	startsTruth := make([]uint32, len(startsIPs))
	for i, ip := range startsIPs {
		converted, err := ipToInt(ip)
		if err != nil {
			t.Errorf("TestGetRanges failed @ ipToInt")
		}
		startsTruth[i] = converted
	}

	endsIPs := []string{
		"10.240.9.255",
		"10.240.219.255",
		"10.240.255.255",
	}
	endsTruth := make([]uint32, len(endsIPs))
	for i, ip := range endsIPs {
		converted, err := ipToInt(ip)
		if err != nil {
			t.Errorf("TestGetRanges failed @ ipToInt")
		}
		endsTruth[i] = converted
	}

	if !reflect.DeepEqual(starts, startsTruth) {
		t.Errorf("TestGetRanges failed @ starts comparison")
	}

	if !reflect.DeepEqual(ends, endsTruth) {
		t.Errorf("TestGetRanges failed @ ends comparison")
	}
}

func TestGetStrCIDR(t *testing.T) {
	strCIDRs := []string{
		"0.0.0.0/16",
		"255.0.1.16/20",
		"10.240.0.0/24",
		"12.144.2.1/31",
		"240.220.10.6/18",
		"11.82.80.0/21",
	}

	var reconstructed []string
	for _, strCIDR := range strCIDRs {
		arrCIDR := strings.Split(strCIDR, "/")
		ip := ipToInt(arrCIDR[0])
		maskNum64, err := strconv.ParseInt(arrCIDR[1], 10, 6)
		if err != nil {
			t.Errorf("TestGetStrCIDR failed @ strconv.ParseUint")
		}
		maskNum := int(maskNum64)
		reconstructed = append(reconstructed, getStrCIDR(ip, maskNum))
	}

	if !reflect.DeepEqual(strCIDRs, reconstructed) {
		t.Errorf("TestGetStrCIDR failed @ strCIDRs comparison")
	}
}