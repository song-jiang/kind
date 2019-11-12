/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"k8s.io/klog"
)

func syncRoute(nodeIP string, podCIDRs []string) error {
	ip := net.ParseIP(nodeIP)
	// return if is not a global unicast address
	// we can't use to route our traffic
	if !ip.IsGlobalUnicast() {
		// TODO fail and handle the error
		return nil
	}
	for _, podCIDR := range podCIDRs {
		// parse subnet
		dst, err := netlink.ParseIPNet(podCIDR)
		if err != nil {
			return err
		}

		// Check if the route exists to the other node's PodCIDR
		routeToDst := netlink.Route{Dst: dst, Gw: ip}
		route, err := netlink.RouteListFiltered(nl.GetIPFamily(ip), &routeToDst, netlink.RT_FILTER_DST)
		if err != nil {
			return err
		}

		// Add route if not present
		if len(route) == 0 {
			klog.Infof("Adding route %v \n", routeToDst)
			if err := netlink.RouteAdd(&routeToDst); err != nil {
				return err
			}
		}
	}
	return nil
}
