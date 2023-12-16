/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package properties

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/incubator-kie-kogito-serverless-operator/api/metadata"
	operatorapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/discovery"
	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/platform/services"
	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/profiles/common/constants"

	"github.com/magiconair/properties"

	"github.com/stretchr/testify/assert"

	"github.com/apache/incubator-kie-kogito-serverless-operator/test"
)

const (
	defaultNamespace  = "default-namespace"
	namespace1        = "namespace1"
	myService1        = "my-service1"
	myService1Address = "http://10.110.90.1:80"
	myService2        = "my-service2"
	myService2Address = "http://10.110.90.2:80"
	myService3        = "my-service3"
	myService3Address = "http://10.110.90.3:80"
)

type mockCatalogService struct {
}

func (c *mockCatalogService) Query(ctx context.Context, uri discovery.ResourceUri, outputFormat string) (string, error) {
	if uri.Scheme == discovery.KubernetesScheme && uri.Namespace == namespace1 && uri.Name == myService1 {
		return myService1Address, nil
	}
	if uri.Scheme == discovery.KubernetesScheme && uri.Name == myService2 && uri.Namespace == defaultNamespace {
		return myService2Address, nil
	}
	if uri.Scheme == discovery.KubernetesScheme && uri.Name == myService3 && uri.Namespace == defaultNamespace && uri.GetPort() == "http-port" {
		return myService3Address, nil
	}
	return "", nil
}

func Test_appPropertyHandler_WithKogitoServiceUrl(t *testing.T) {
	workflow := test.GetBaseSonataFlow("default")
	props, err := ImmutableApplicationProperties(workflow, nil)
	assert.NoError(t, err)
	assert.Contains(t, props, constants.KogitoServiceURLProperty)
	assert.Contains(t, props, "http://"+workflow.Name+"."+workflow.Namespace)
}

const (
	defaultNS = "default"
)

func Test_appPropertyHandler_WithJobServiceInDevProfile(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceWithPostgreSQLInDevProfile(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceWithPostgreSQLInProdProfile(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.ProdProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Set("kogito.events.processinstances.enabled", "false")
	expected.Set("quarkus.datasource.reactive.url", "postgresql://foo.default:5432/postgres?search_path=job-service")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceAndDataIndexEnabledInDevProfile(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceAndDataIndexEnabledInProdProfile(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.ProdProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processinstances.enabled", "true")
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Set("mp.messaging.outgoing.kogito-processinstances-events.url", "http://sonataflow-platform-data-index-service.default/processes")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceAndDataIndexWithPostgreSQLDevMode(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceAndDataIndexWithPostgreSQLProdMode(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.ProdProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Set("kogito.events.processinstances.enabled", "true")
	expected.Set("mp.messaging.outgoing.kogito-processinstances-events.url", "http://sonataflow-platform-data-index-service.default/processes")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceWithPostgreSQLDevMode(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithJobServiceWithPostgreSQLProdMode(t *testing.T) {
	workflow := test.GetBaseSonataFlow(defaultNS)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.ProdProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = defaultNS
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
				Persistence: &operatorapi.PersistenceOptions{
					PostgreSql: &operatorapi.PersistencePostgreSql{
						ServiceRef: &operatorapi.PostgreSqlServiceOptions{
							Name:      "foo",
							Namespace: defaultNS,
						},
						JdbcUrl: "jdbc:postgresql://foo.default:5432/postgres?currentSchema=job-service",
					},
				},
			},
		},
	}
	p, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	props, err := properties.LoadString(p.Build())
	assert.NoError(t, err)
	expected := properties.NewProperties()
	expected.Set("kogito.events.processdefinitions.enabled", "false")
	expected.Set("kogito.events.usertasks.enabled", "false")
	expected.Set("kogito.events.variables.enabled", "false")
	expected.Set("kogito.service.url", "http://greeting.default")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.connector", "quarkus-http")
	expected.Set("mp.messaging.outgoing.kogito-job-service-job-request-events.url", "http://sonataflow-platform-jobs-service.default/v2/jobs/events")
	expected.Set("org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	expected.Set("quarkus.devservices.enabled", "false")
	expected.Set("quarkus.http.host", "0.0.0.0")
	expected.Set("quarkus.http.port", "8080")
	expected.Set("quarkus.kogito.devservices.enabled", "false")
	expected.Set("kogito.events.processinstances.enabled", "true")
	expected.Set("mp.messaging.outgoing.kogito-processinstances-events.url", "http://sonataflow-platform-data-index-service.default/processes")
	expected.Set("quarkus.datasource.reactive.url", "postgresql://foo.default:5432/postgres?search_path=job-service")
	expected.Sort()
	assert.Equal(t, expected, props)
}

func Test_appPropertyHandler_WithUserPropertiesWithNoUserOverrides(t *testing.T) {
	//just add some user provided properties, no overrides.
	userProperties := "property1=value1\nproperty2=value2"
	workflow := test.GetBaseSonataFlow("default")
	props, err := NewAppPropertyHandler(workflow, nil)
	assert.NoError(t, err)
	generatedProps, propsErr := properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 11, len(generatedProps.Keys()))
	assert.Equal(t, "value1", generatedProps.GetString("property1", ""))
	assert.Equal(t, "value2", generatedProps.GetString("property2", ""))
	assert.Equal(t, "http://greeting.default", generatedProps.GetString("kogito.service.url", ""))
	assert.Equal(t, "8080", generatedProps.GetString("quarkus.http.port", ""))
	assert.Equal(t, "0.0.0.0", generatedProps.GetString("quarkus.http.host", ""))
	assert.Equal(t, "false", generatedProps.GetString("org.kie.kogito.addons.knative.eventing.health-enabled", ""))
	assert.Equal(t, "false", generatedProps.GetString("quarkus.devservices.enabled", ""))
	assert.Equal(t, "false", generatedProps.GetString("quarkus.kogito.devservices.enabled", ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoProcessDefinitionsEnabled, ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoEventsUserTaskEnabled, ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoEventsVariablesEnabled, ""))
}

func Test_appPropertyHandler_WithUserPropertiesWithServiceDiscovery(t *testing.T) {
	//just add some user provided properties, no overrides.
	userProperties := "property1=value1\nproperty2=value2\n"
	//add some user properties that requires service discovery
	userProperties = userProperties + "service1=${kubernetes:services.v1/namespace1/my-service1}\n"
	userProperties = userProperties + "service2=${kubernetes:services.v1/my-service2}\n"

	workflow := test.GetBaseSonataFlow(defaultNamespace)
	props, err := NewAppPropertyHandler(workflow, nil)
	assert.NoError(t, err)
	generatedProps, propsErr := properties.LoadString(props.
		WithUserProperties(userProperties).
		WithServiceDiscovery(context.TODO(), &mockCatalogService{}).
		Build())
	generatedProps.DisableExpansion = true
	assert.NoError(t, propsErr)
	assert.Equal(t, 15, len(generatedProps.Keys()))
	assertHasProperty(t, generatedProps, "property1", "value1")
	assertHasProperty(t, generatedProps, "property2", "value2")

	assertHasProperty(t, generatedProps, "service1", "${kubernetes:services.v1/namespace1/my-service1}")
	assertHasProperty(t, generatedProps, "service2", "${kubernetes:services.v1/my-service2}")
	//org.kie.kogito.addons.discovery.kubernetes\:services.v1\/usecase1º/my-service1 below we use the unescaped vale because the properties.LoadString removes them.
	assertHasProperty(t, generatedProps, "org.kie.kogito.addons.discovery.kubernetes:services.v1/namespace1/my-service1", myService1Address)
	//org.kie.kogito.addons.discovery.kubernetes\:services.v1\/my-service2 below we use the unescaped vale because the properties.LoadString removes them.
	assertHasProperty(t, generatedProps, "org.kie.kogito.addons.discovery.kubernetes:services.v1/my-service2", myService2Address)

	assertHasProperty(t, generatedProps, "kogito.service.url", fmt.Sprintf("http://greeting.%s", defaultNamespace))
	assertHasProperty(t, generatedProps, "quarkus.http.port", "8080")
	assertHasProperty(t, generatedProps, "quarkus.http.host", "0.0.0.0")
	assertHasProperty(t, generatedProps, "org.kie.kogito.addons.knative.eventing.health-enabled", "false")
	assertHasProperty(t, generatedProps, "quarkus.devservices.enabled", "false")
	assertHasProperty(t, generatedProps, "quarkus.kogito.devservices.enabled", "false")
	assertHasProperty(t, generatedProps, constants.KogitoProcessDefinitionsEnabled, "false")
	assertHasProperty(t, generatedProps, constants.KogitoEventsUserTaskEnabled, "false")
	assertHasProperty(t, generatedProps, constants.KogitoEventsVariablesEnabled, "false")
}

func Test_generateDiscoveryProperties(t *testing.T) {

	catalogService := &mockCatalogService{}

	propertiesContent := "property1=value1\n"
	propertiesContent = propertiesContent + "property2=${value2}\n"
	propertiesContent = propertiesContent + "service1=${kubernetes:services.v1/namespace1/my-service1}\n"
	propertiesContent = propertiesContent + "service2=${kubernetes:services.v1/my-service2}\n"
	propertiesContent = propertiesContent + "service3=${kubernetes:services.v1/my-service3?port=http-port}\n"

	propertiesContent = propertiesContent + "non_service4=${kubernetes:--kaka}"

	props := properties.MustLoadString(propertiesContent)
	result := generateDiscoveryProperties(context.TODO(), catalogService, props, &operatorapi.SonataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: "helloworld", Namespace: defaultNamespace},
	})

	assert.Equal(t, result.Len(), 3)
	assertHasProperty(t, result, "org.kie.kogito.addons.discovery.kubernetes\\:services.v1\\/namespace1\\/my-service1", myService1Address)
	assertHasProperty(t, result, "org.kie.kogito.addons.discovery.kubernetes\\:services.v1\\/my-service2", myService2Address)
	assertHasProperty(t, result, "org.kie.kogito.addons.discovery.kubernetes\\:services.v1\\/my-service3?port\\=http-port", myService3Address)
}

func assertHasProperty(t *testing.T, props *properties.Properties, expectedProperty string, expectedValue string) {
	value, ok := props.Get(expectedProperty)
	assert.True(t, ok, "Property %s, is not present as expected.", expectedProperty)
	assert.Equal(t, expectedValue, value, "Expected value for property: %s, is: %s but current value is: %s", expectedProperty, expectedValue, value)
}

func Test_generateMicroprofileServiceCatalogProperty(t *testing.T) {

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:services.v1/namespace1/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:services.v1\\/namespace1\\/financial-service")

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:services.v1/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:services.v1\\/financial-service")

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:pods.v1/namespace1/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:pods.v1\\/namespace1\\/financial-service")

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:pods.v1/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:pods.v1\\/financial-service")

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:deployments.v1.apps/namespace1/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:deployments.v1.apps\\/namespace1\\/financial-service")

	doTestGenerateMicroprofileServiceCatalogProperty(t, "kubernetes:deployments.v1.apps/financial-service",
		"org.kie.kogito.addons.discovery.kubernetes\\:deployments.v1.apps\\/financial-service")
}

func doTestGenerateMicroprofileServiceCatalogProperty(t *testing.T, serviceUri string, expectedProperty string) {
	mpProperty := generateMicroprofileServiceCatalogProperty(serviceUri)
	assert.Equal(t, mpProperty, expectedProperty, "expected microprofile service catalog property for serviceUri: %s, is %s, but the returned value was: %s", serviceUri, expectedProperty, mpProperty)
}

func Test_appPropertyHandler_WithServicesWithUserOverrides(t *testing.T) {
	//try to override kogito.service.url and quarkus.http.port
	userProperties := "property1=value1\nproperty2=value2\nquarkus.http.port=9090\nkogito.service.url=http://myUrl.override.com\nquarkus.http.port=9090"
	ns := "default"
	workflow := test.GetBaseSonataFlow(ns)
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.DevProfile)})
	enabled := true
	platform := test.GetBasePlatform()
	platform.Namespace = ns
	platform.Spec = operatorapi.SonataFlowPlatformSpec{
		Services: operatorapi.ServicesPlatformSpec{
			DataIndex: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
			JobService: &operatorapi.ServiceSpec{
				Enabled: &enabled,
			},
		},
	}

	props, err := NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr := properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 13, len(generatedProps.Keys()))
	assert.Equal(t, "value1", generatedProps.GetString("property1", ""))
	assert.Equal(t, "value2", generatedProps.GetString("property2", ""))

	//kogito.service.url takes the user provided value since it's a default mutable property.
	assert.Equal(t, "http://myUrl.override.com", generatedProps.GetString("kogito.service.url", ""))
	//quarkus.http.port remains with the default value since it's immutable.
	assert.Equal(t, "8080", generatedProps.GetString("quarkus.http.port", ""))
	assert.Equal(t, "0.0.0.0", generatedProps.GetString("quarkus.http.host", ""))
	assert.Equal(t, "false", generatedProps.GetString("org.kie.kogito.addons.knative.eventing.health-enabled", ""))
	assert.Equal(t, "false", generatedProps.GetString("quarkus.devservices.enabled", ""))
	assert.Equal(t, "false", generatedProps.GetString("quarkus.kogito.devservices.enabled", ""))
	assert.Equal(t, "", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://sonataflow-platform-jobs-service.default/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoProcessDefinitionsEnabled, ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoEventsUserTaskEnabled, ""))
	assert.Equal(t, "false", generatedProps.GetString(constants.KogitoEventsVariablesEnabled, ""))

	// prod profile enables config of outgoing events url
	workflow.SetAnnotations(map[string]string{metadata.Profile: string(metadata.ProdProfile)})
	props, err = NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 15, len(generatedProps.Keys()))
	assert.Equal(t, "http://"+platform.Name+"-"+constants.DataIndexServiceName+"."+platform.Namespace+"/processes", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://"+platform.Name+"-"+constants.JobServiceName+"."+platform.Namespace+"/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceDataSourceReactiveURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEvents, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEventsURL, ""))

	// disabling data index bypasses config of outgoing events url
	platform.Spec.Services.DataIndex.Enabled = nil
	props, err = NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 14, len(generatedProps.Keys()))
	assert.Equal(t, "", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://"+platform.Name+"-"+constants.JobServiceName+"."+platform.Namespace+"/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEvents, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEventsURL, ""))

	// check that service app properties are being properly set
	js := services.NewJobService(platform)
	p, err := NewServiceAppPropertyHandler(platform, js)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(p.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 10, len(generatedProps.Keys()))
	assert.Equal(t, "false", generatedProps.GetString(constants.KafkaSmallRyeHealthProperty, ""))
	assert.Equal(t, "value1", generatedProps.GetString("property1", ""))
	assert.Equal(t, "value2", generatedProps.GetString("property2", ""))
	//quarkus.http.port remains with the default value since it's immutable.
	assert.Equal(t, "8080", generatedProps.GetString("quarkus.http.port", ""))

	// disabling job service bypasses config of outgoing events url
	platform.Spec.Services.JobService.Enabled = nil
	props, err = NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 13, len(generatedProps.Keys()))
	assert.Equal(t, "", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://sonataflow-platform-jobs-service.default/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceDataSourceReactiveURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEvents, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEventsURL, ""))

	// check that the reactive URL is generated from the postgreSQL JDBC URL when not provided
	platform.Spec.Services.JobService = &operatorapi.ServiceSpec{
		Enabled: &enabled,
		Persistence: &operatorapi.PersistenceOptions{
			PostgreSql: &operatorapi.PersistencePostgreSql{
				ServiceRef: &operatorapi.PostgreSqlServiceOptions{
					Name: "jobs-service",
				},
			},
		},
	}
	props, err = NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 15, len(generatedProps.Keys()))
	assert.Equal(t, "", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://"+platform.Name+"-"+constants.JobServiceName+"."+platform.Namespace+"/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "postgresql://jobs-service.default:5432/sonataflow?search_path=sonataflow-platform-jobs-service", generatedProps.GetString(constants.JobServiceDataSourceReactiveURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEvents, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEventsURL, ""))

	// check that the reactive URL is generated from the postgreSQL JDBC URL when provided
	platform.Spec.Services.JobService = &operatorapi.ServiceSpec{
		Enabled: &enabled,
		Persistence: &operatorapi.PersistenceOptions{
			PostgreSql: &operatorapi.PersistencePostgreSql{
				JdbcUrl: "jdbc:postgresql://timeouts-showcase-database:5432/postgres?currentSchema=jobs-service",
			},
		},
	}
	props, err = NewAppPropertyHandler(workflow, platform)
	assert.NoError(t, err)
	generatedProps, propsErr = properties.LoadString(props.WithUserProperties(userProperties).Build())
	assert.NoError(t, propsErr)
	assert.Equal(t, 15, len(generatedProps.Keys()))
	assert.Equal(t, "", generatedProps.GetString(constants.DataIndexServiceURLProperty, ""))
	assert.Equal(t, "http://"+platform.Name+"-"+constants.JobServiceName+"."+platform.Namespace+"/v2/jobs/events", generatedProps.GetString(constants.JobServiceRequestEventsURL, ""))
	assert.Equal(t, "postgresql://timeouts-showcase-database:5432/postgres?search_path=jobs-service", generatedProps.GetString(constants.JobServiceDataSourceReactiveURL, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEvents, ""))
	assert.Equal(t, "", generatedProps.GetString(constants.JobServiceStatusChangeEventsURL, ""))

}