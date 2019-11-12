/*
Copyright 2018 The Kubernetes Authors.

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

package config

import (
	"testing"

	"sigs.k8s.io/kind/pkg/errors"
)

func TestClusterValidate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name         string
		Cluster      Cluster
		ExpectErrors int
	}{
		{
			Name: "Defaulted",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				return c
			}(),
		},
		{
			Name: "multiple valid nodes",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Nodes = append(c.Nodes, newDefaultedNode(WorkerRole), newDefaultedNode(WorkerRole))
				return c
			}(),
		},
		{
			Name: "default IPv6",
			Cluster: func() Cluster {
				c := Cluster{}
				c.Networking.IPFamily = IPv6Family
				SetDefaultsCluster(&c)
				return c
			}(),
		},
		{
			Name: "bogus podSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "aa"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bogus serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.ServiceSubnet = "aa"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bogus apiServerPort",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.APIServerPort = 9999999
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bogus serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.ServiceSubnet = "aa"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "invalid number of podSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "192.168.0.2/24,2.2.2.0/24"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "valid dual stack podSubnet and serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "192.168.0.2/24,fd00:1::/25"
				c.Networking.ServiceSubnet = "192.168.0.2/24,fd00:1::/25"
				c.Networking.IPFamily = "DualStack"
				return c
			}(),
			ExpectErrors: 0,
		},
		{
			Name: "valid dual stack podSubnet and bad serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "192.168.0.2/24,fd00:1::/25"
				c.Networking.ServiceSubnet = "192.168.0.2/24"
				c.Networking.IPFamily = "DualStack"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bad dual stack podSubnet and valid serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "fd00:1::/25"
				c.Networking.ServiceSubnet = "fd00:1::/25,192.168.0.2/24"
				c.Networking.IPFamily = "DualStack"
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bad dual stack podSubnet and serviceSubnet",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Networking.PodSubnet = "192.168.0.2/24,2.2.2.0/25"
				c.Networking.ServiceSubnet = "192.168.0.2/24,2.2.2.0/25"
				c.Networking.IPFamily = "DualStack"
				return c
			}(),
			ExpectErrors: 2,
		},
		{
			Name: "missing control-plane",
			Cluster: func() Cluster {
				c := Cluster{}
				SetDefaultsCluster(&c)
				c.Nodes = []Node{}
				return c
			}(),
			ExpectErrors: 1,
		},
		{
			Name: "bogus node",
			Cluster: func() Cluster {
				c := Cluster{}
				n, n2 := Node{}, Node{}
				n.Role = "bogus"
				c.Nodes = []Node{n, n2}
				SetDefaultsCluster(&c)
				return c
			}(),
			ExpectErrors: 1,
		},
	}

	for _, tc := range cases {
		tc := tc //capture loop variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := tc.Cluster.Validate()
			// the error can be:
			// - nil, in which case we should expect no errors or fail
			if err == nil {
				if tc.ExpectErrors != 0 {
					t.Error("received no errors but expected errors for case")
				}
				return
			}
			// get the list of errors
			errs := errors.Errors(err)
			if errs == nil {
				errs = []error{err}
			}
			// we expect a certain number of errors
			if len(errs) != tc.ExpectErrors {
				t.Errorf("expected %d errors but got len(%v) = %d", tc.ExpectErrors, errs, len(errs))
			}
		})
	}
}

func newDefaultedNode(role NodeRole) Node {
	n := Node{
		Role:  role,
		Image: "myImage:latest",
	}
	SetDefaultsNode(&n)
	return n
}

func TestNodeValidate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		TestName     string
		Node         Node
		ExpectErrors int
	}{
		{
			TestName:     "Canonical node",
			Node:         newDefaultedNode(ControlPlaneRole),
			ExpectErrors: 0,
		},
		{
			TestName:     "Canonical node 2",
			Node:         newDefaultedNode(WorkerRole),
			ExpectErrors: 0,
		},
		{
			TestName: "Empty image field",
			Node: func() Node {
				cfg := newDefaultedNode(ControlPlaneRole)
				cfg.Image = ""
				return cfg
			}(),
			ExpectErrors: 1,
		},
		{
			TestName: "Empty role field",
			Node: func() Node {
				cfg := newDefaultedNode(ControlPlaneRole)
				cfg.Role = ""
				return cfg
			}(),
			ExpectErrors: 1,
		},
		{
			TestName: "Unknown role field",
			Node: func() Node {
				cfg := newDefaultedNode(ControlPlaneRole)
				cfg.Role = "ssss"
				return cfg
			}(),
			ExpectErrors: 1,
		},
		{
			TestName: "Invalid ContainerPort",
			Node: func() Node {
				cfg := newDefaultedNode(ControlPlaneRole)
				cfg.ExtraPortMappings = []PortMapping{
					{
						ContainerPort: 999999999,
						HostPort:      8080,
					},
				}
				return cfg
			}(),
			ExpectErrors: 1,
		},
		{
			TestName: "Invalid HostPort",
			Node: func() Node {
				cfg := newDefaultedNode(ControlPlaneRole)
				cfg.ExtraPortMappings = []PortMapping{
					{
						ContainerPort: 8080,
						HostPort:      999999999,
					},
				}
				return cfg
			}(),
			ExpectErrors: 1,
		},
	}

	for _, tc := range cases {
		tc := tc //capture loop variable
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			err := tc.Node.Validate()
			// the error can be:
			// - nil, in which case we should expect no errors or fail
			if err == nil {
				if tc.ExpectErrors != 0 {
					t.Error("received no errors but expected errors for case")
				}
				return
			}
			// get the list of errors
			errs := errors.Errors(err)
			if errs == nil {
				errs = []error{err}
			}
			// we expect a certain number of errors
			if len(errs) != tc.ExpectErrors {
				t.Errorf("expected %d errors but got len(%v) = %d", tc.ExpectErrors, errs, len(errs))
			}
		})
	}
}
