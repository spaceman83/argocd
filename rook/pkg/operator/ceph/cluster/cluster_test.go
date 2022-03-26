/*
Copyright 2020 The Rook Authors. All rights reserved.

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

// Package cluster to manage a Ceph cluster.
package cluster

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/daemon/ceph/client"
	cephclient "github.com/rook/rook/pkg/daemon/ceph/client"
	cephver "github.com/rook/rook/pkg/operator/ceph/version"
	testop "github.com/rook/rook/pkg/operator/test"
	exectest "github.com/rook/rook/pkg/util/exec/test"
	"github.com/stretchr/testify/assert"
)

func TestPreClusterStartValidation(t *testing.T) {
	type args struct {
		cluster *cluster
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no settings", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), Spec: &cephv1.ClusterSpec{}, context: &clusterd.Context{Clientset: testop.New(t, 3)}}}, false},
		{"even mons", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{Count: 2}}}}, false},
		{"missing stretch zones", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{StretchCluster: &cephv1.StretchClusterSpec{Zones: []cephv1.StretchClusterZoneSpec{
			{Name: "a"},
		}}}}}}, true},
		{"missing arbiter", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{StretchCluster: &cephv1.StretchClusterSpec{Zones: []cephv1.StretchClusterZoneSpec{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		}}}}}}, true},
		{"missing zone name", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{StretchCluster: &cephv1.StretchClusterSpec{Zones: []cephv1.StretchClusterZoneSpec{
			{Arbiter: true},
			{Name: "b"},
			{Name: "c"},
		}}}}}}, true},
		{"valid stretch cluster", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{Count: 3, StretchCluster: &cephv1.StretchClusterSpec{Zones: []cephv1.StretchClusterZoneSpec{
			{Name: "a", Arbiter: true},
			{Name: "b"},
			{Name: "c"},
		}}}}}}, false},
		{"not enough stretch nodes", args{&cluster{ClusterInfo: client.AdminTestClusterInfo("rook-ceph"), context: &clusterd.Context{Clientset: testop.New(t, 3)}, Spec: &cephv1.ClusterSpec{Mon: cephv1.MonSpec{Count: 5, StretchCluster: &cephv1.StretchClusterSpec{Zones: []cephv1.StretchClusterZoneSpec{
			{Name: "a", Arbiter: true},
			{Name: "b"},
			{Name: "c"},
		}}}}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := preClusterStartValidation(tt.args.cluster); (err != nil) != tt.wantErr {
				t.Errorf("ClusterController.preClusterStartValidation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreMonChecks(t *testing.T) {
	executor := &exectest.MockExecutor{}
	context := &clusterd.Context{Executor: executor}
	setSkipSanity := false
	unsetSkipSanity := false
	executor.MockExecuteCommandWithTimeout = func(timeout time.Duration, command string, args ...string) (string, error) {
		logger.Infof("Command: %s %v", command, args)
		if args[0] == "config" {
			if args[1] == "set" {
				setSkipSanity = true
				assert.Equal(t, "mon", args[2])
				assert.Equal(t, "mon_mds_skip_sanity", args[3])
				assert.Equal(t, "1", args[4])
				return "", nil
			}
			if args[1] == "rm" {
				unsetSkipSanity = true
				assert.Equal(t, "mon", args[2])
				assert.Equal(t, "mon_mds_skip_sanity", args[3])
				return "", nil
			}
		}
		return "", errors.Errorf("unexpected ceph command %q", args)
	}
	c := cluster{context: context, ClusterInfo: cephclient.AdminTestClusterInfo("cluster")}

	t.Run("no upgrade", func(t *testing.T) {
		v := cephver.CephVersion{Major: 16, Minor: 2, Extra: 7}
		c.isUpgrade = false
		err := c.preMonStartupActions(v)
		assert.NoError(t, err)
		assert.False(t, setSkipSanity)
		assert.False(t, unsetSkipSanity)
	})

	t.Run("upgrade below version", func(t *testing.T) {
		setSkipSanity = false
		unsetSkipSanity = false
		v := cephver.CephVersion{Major: 16, Minor: 2, Extra: 6}
		c.isUpgrade = true
		err := c.preMonStartupActions(v)
		assert.NoError(t, err)
		assert.False(t, setSkipSanity)
		assert.False(t, unsetSkipSanity)
	})

	t.Run("upgrade to applicable version", func(t *testing.T) {
		setSkipSanity = false
		unsetSkipSanity = false
		v := cephver.CephVersion{Major: 16, Minor: 2, Extra: 7}
		c.isUpgrade = true
		err := c.preMonStartupActions(v)
		assert.NoError(t, err)
		assert.True(t, setSkipSanity)
		assert.False(t, unsetSkipSanity)

		// This will be called during the post mon checks
		err = c.skipMDSSanityChecks(false)
		assert.NoError(t, err)
		assert.True(t, unsetSkipSanity)
	})

	t.Run("upgrade to quincy", func(t *testing.T) {
		setSkipSanity = false
		unsetSkipSanity = false
		v := cephver.CephVersion{Major: 17, Minor: 2, Extra: 0}
		c.isUpgrade = true
		err := c.preMonStartupActions(v)
		assert.NoError(t, err)
		assert.False(t, setSkipSanity)
		assert.False(t, unsetSkipSanity)
	})
}
