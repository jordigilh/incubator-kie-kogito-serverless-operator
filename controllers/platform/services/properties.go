// Copyright 2023 Apache Software Foundation (ASF)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"fmt"
	"net/url"
	"strings"

	operatorapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/profiles"
	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/profiles/common/constants"

	"github.com/magiconair/properties"
)

func generateReactiveURL(postgresSpec *operatorapi.PersistencePostgreSql, schema string, namespace string, dbName string, port int) (string, error) {
	if len(postgresSpec.JdbcUrl) > 0 {
		s := strings.TrimLeft(postgresSpec.JdbcUrl, "jdbc:")
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		ret := fmt.Sprintf("%s://", u.Scheme)
		if len(u.User.Username()) > 0 {
			p, ok := u.User.Password()
			if ok {
				ret = fmt.Sprintf("%s%s:%s@", ret, u.User.Username(), p)
			}
		}
		ret = fmt.Sprintf("%s%s%s", ret, u.Host, u.Path)
		kv, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return "", err
		}
		var spv string
		if v, ok := kv["search_path"]; ok {
			for _, val := range v {
				if len(val) != 0 {
					spv = v[0]
				}
			}
		} else if v, ok := kv["currentSchema"]; ok {
			for _, val := range v {
				if len(val) != 0 {
					spv = v[0]
				}
			}
		}
		if len(spv) > 0 {
			return fmt.Sprintf("%s?search_path=%s", ret, spv), nil
		}
		return ret, nil
	}
	databaseSchema := schema
	if len(postgresSpec.ServiceRef.DatabaseSchema) > 0 {
		databaseSchema = postgresSpec.ServiceRef.DatabaseSchema
	}
	databaseNamespace := namespace
	if len(postgresSpec.ServiceRef.Namespace) > 0 {
		databaseNamespace = postgresSpec.ServiceRef.Namespace
	}
	dataSourcePort := port
	if postgresSpec.ServiceRef.Port != nil {
		dataSourcePort = *postgresSpec.ServiceRef.Port
	}
	databaseName := dbName
	if len(postgresSpec.ServiceRef.DatabaseName) > 0 {
		databaseName = postgresSpec.ServiceRef.DatabaseName
	}
	return fmt.Sprintf("%s://%s:%d/%s?search_path=%s", constants.PersistenceTypePostgreSQL, postgresSpec.ServiceRef.Name+"."+databaseNamespace, dataSourcePort, databaseName, databaseSchema), nil
}

// GenerateDataIndexApplicationProperties returns the application properties required for the Data Index Service to work when deployed in a production profile
// and data index' service's spec field `Enabled` set to true
func GenerateDataIndexApplicationProperties(workflow *operatorapi.SonataFlow, platform *operatorapi.SonataFlowPlatform) (*properties.Properties, error) {
	var err error
	props := properties.NewProperties()
	if workflow != nil && !profiles.IsDevProfile(workflow) && dataIndexEnabled(platform) {
		di := NewDataIndexService(platform)
		props, err = di.GenerateWorkflowProperties()
		if err != nil {
			return nil, err
		}
		props.Sort()
	}
	return props, nil
}

// GenerateJobServiceApplicationProperties returns the application properties required for the Job Service to work in a production profile and job service's
// service's spec field `Enabled` set to true
func GenerateJobServiceApplicationProperties(workflow *operatorapi.SonataFlow, platform *operatorapi.SonataFlowPlatform) (*properties.Properties, error) {
	js := NewJobService(platform)
	props := properties.NewProperties()
	props.Set(constants.JobServiceRequestEventsConnector, constants.QuarkusHTTP)
	props.Set(
		constants.JobServiceRequestEventsURL, fmt.Sprintf("%s://%s.%s/v2/jobs/events", constants.JobServiceURLProtocol, js.GetServiceName(), platform.Namespace))
	if workflow != nil && !profiles.IsDevProfile(workflow) && jobServiceEnabled(platform) {
		p, err := js.GenerateWorkflowProperties()
		if err != nil {
			return nil, err
		}
		props.Merge(p)
	}
	props.Sort()
	return props, nil
}

func dataIndexEnabled(platform *operatorapi.SonataFlowPlatform) bool {
	return platform != nil && platform.Spec.Services.DataIndex != nil &&
		platform.Spec.Services.DataIndex.Enabled != nil && *platform.Spec.Services.DataIndex.Enabled
}

func jobServiceEnabled(platform *operatorapi.SonataFlowPlatform) bool {
	return platform != nil && platform.Spec.Services.JobService != nil && platform.Spec.Services.JobService.Enabled != nil && *platform.Spec.Services.JobService.Enabled
}