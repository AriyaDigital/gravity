/*
Copyright 2018 Gravitational, Inc.

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

package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gravitational/gravity/lib/constants"
	"github.com/gravitational/gravity/lib/defaults"

	teleservices "github.com/gravitational/teleport/lib/services"
	teleutils "github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
)

// EnvironmentVariables defines the environment variables resource.
// It allows to override environment variables on each node in the cluster.
// There is only a single instance of the resource in a cluster
type EnvironmentVariables interface {
	// Resource provides common resource methods
	teleservices.Resource
	// CheckAndSetDefaults validates this resource and sets defaults
	CheckAndSetDefaults() error
	SetKeyValues(map[string]string)
	GetKeyValues() map[string]string
}

// NewEnvironment creates a new instance of the resource
func NewEnvironment(kvs map[string]string) *EnvironmentV1 {
	return &EnvironmentV1{
		Kind:    KindEnvironment,
		Version: "v1",
		Metadata: teleservices.Metadata{
			Name:      constants.ClusterEnvironmentMap,
			Namespace: defaults.Namespace,
		},
		Spec: EnvironmentSpec{
			KeyValues: kvs,
		},
	}
}

// EnvironmentV1 describes the environment variable resource
type EnvironmentV1 struct {
	// Kind is a resource kind
	Kind string `json:"kind"`
	// Version is a resource version
	Version string `json:"version"`
	// Metadata specifies resource metadata
	Metadata teleservices.Metadata `json:"metadata"`
	// Spec defines the resource
	Spec EnvironmentSpec `json:"spec"`
}

// GetName returns the name of the resource name
func (r *EnvironmentV1) GetName() string {
	return r.Metadata.Name
}

// SetName resets the resource name to the specified value
func (r *EnvironmentV1) SetName(name string) {
	r.Metadata.Name = name
}

// GetMetadata returns resource metadata
func (r *EnvironmentV1) GetMetadata() teleservices.Metadata {
	return r.Metadata
}

// SetExpiry resets expiration time to the specified value
func (r *EnvironmentV1) SetExpiry(expires time.Time) {
	r.Metadata.SetExpiry(expires)
}

// Expires returns expiration time
func (r *EnvironmentV1) Expiry() time.Time {
	return r.Metadata.Expiry()
}

// SetTTL resets the resources's time to live to the specified value
// using given clock implementation
func (r *EnvironmentV1) SetTTL(clock clockwork.Clock, ttl time.Duration) {
	r.Metadata.SetTTL(clock, ttl)
}

// GetKeyValues returns the variables from this environment
func (r *EnvironmentV1) GetKeyValues() map[string]string {
	return r.Spec.KeyValues
}

// SetKeyValues returns the variables from this environment
func (r *EnvironmentV1) SetKeyValues(kvs map[string]string) {
	r.Spec.KeyValues = kvs
}

// CheckAndSetDefaults validates this resource and sets defaults
func (r *EnvironmentV1) CheckAndSetDefaults() error {
	return nil
}

// UnmarshalEnvironmentVariables unmarshals the resource from JSON given with data
func UnmarshalEnvironmentVariables(data []byte) (EnvironmentVariables, error) {
	if len(data) == 0 {
		return &EnvironmentV1{}, nil
	}
	jsonData, err := teleutils.ToJSON(data)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var hdr teleservices.ResourceHeader
	err = json.Unmarshal(jsonData, &hdr)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	switch hdr.Version {
	case "v1":
		var env EnvironmentV1
		err := teleutils.UnmarshalWithSchema(GetEnvironmentSpecSchema(), &env, jsonData)
		if err != nil {
			return nil, trace.BadParameter(err.Error())
		}
		if err := env.Metadata.CheckAndSetDefaults(); err != nil {
			return nil, trace.Wrap(err)
		}
		return &env, nil
	}
	return nil, trace.BadParameter(
		"%v resource version %q is not supported", KindEnvironment, hdr.Version)
}

// MarshalEnvironment marshals this resource as JSON
func MarshalEnvironment(env EnvironmentVariables, opts ...teleservices.MarshalOption) ([]byte, error) {
	return json.Marshal(env)
}

// EnvironmentSpec defines the environment variable resource
type EnvironmentSpec struct {
	// KeyValues specifies the environment
	KeyValues map[string]string `json:"data"`
}

// EnvironmentSpecSchema is JSON schema for the environment variables resource
const EnvironmentSpecSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["data"],
  "properties": {
    "data": {"type": ["object", "null"]}
  }
}`

// GetEnvironmentSpecSchema returns the formatted JSON schema for the environment
// variables resource
func GetEnvironmentSpecSchema() string {
	return fmt.Sprintf(teleservices.V2SchemaTemplate, teleservices.MetadataSchema,
		EnvironmentSpecSchema, "")
}
