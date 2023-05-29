// Copyright 2023 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"context"
	"os"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-serverless-operator/container-builder/api"
	"github.com/kiegroup/kogito-serverless-operator/container-builder/util/test"
)

func TestNewBuild(t *testing.T) {
	ns := "test"
	c, err := test.NewFakeClient()
	assert.NoError(t, err)

	dockerFile, err := os.ReadFile("testdata/Dockerfile")
	assert.NoError(t, err)

	workflowDefinition, err := os.ReadFile("testdata/greetings.sw.json")
	assert.NoError(t, err)

	platform := api.PlatformContainerBuild{
		ObjectReference: api.ObjectReference{
			Namespace: ns,
			Name:      "testPlatform",
		},
		Spec: api.PlatformContainerBuildSpec{
			BuildStrategy:   api.ContainerBuildStrategyPod,
			PublishStrategy: api.PlatformBuildPublishStrategyKaniko,
			Timeout:         &metav1.Duration{Duration: 5 * time.Minute},
		},
	}
	// create the new build, schedule
	build, err := NewBuild(ContainerBuilderInfo{FinalImageName: "quay.io/kiegroup/buildexample:latest", BuildUniqueName: "build1", Platform: platform}).
		WithClient(c).
		WithResource("Dockerfile", dockerFile).
		WithResource("greetings.sw.json", workflowDefinition).
		Schedule()

	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.Equal(t, api.ContainerBuildPhaseScheduling, build.Status.Phase)

	build, err = FromBuild(build).WithClient(c).Reconcile()
	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.Equal(t, api.ContainerBuildPhasePending, build.Status.Phase)

	// The status won't change since FakeClient won't set the status upon creation, since we don't have a controller :)
	build, err = FromBuild(build).WithClient(c).Reconcile()
	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.Equal(t, api.ContainerBuildPhasePending, build.Status.Phase)

	podName := buildPodName(build)
	pod := &v1.Pod{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: ns}, pod)
	assert.NoError(t, err)
	assert.NotNil(t, pod)
	assert.Len(t, pod.Spec.Volumes, 1)
}
