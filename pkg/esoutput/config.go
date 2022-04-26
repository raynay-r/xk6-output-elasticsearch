/*
 * Licensed to Elasticsearch B.V. under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Elasticsearch B.V. licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package esoutput

import (
	"encoding/json"
	"time"

	"github.com/kubernetes/helm/pkg/strvals"
	"go.k6.io/k6/lib/types"
	"gopkg.in/guregu/null.v3"
)

const (
	defaultFlushPeriod = time.Second
)

type Config struct {
	Url     null.String `json:"url" envconfig:"K6_ELASTICSEARCH_URL"`
	CloudID null.String `json:"cloud-id"  envconfig:"K6_ELASTICSEARCH_CLOUD_ID"`
	CACert  null.String `json:"caCertFile" envconfig:"K6_ELASTICSEARCH_CA_CERT_FILE"`

	User     null.String `json:"user" envconfig:"K6_ELASTICSEARCH_USER"`
	Password null.String `json:"password" envconfig:"K6_ELASTICSEARCH_PASSWORD"`

	FlushPeriod types.NullDuration `json:"flushPeriod" envconfig:"K6_ELASTICSEARCH_FLUSH_PERIOD"`
}

func NewConfig() Config {
	return Config{
		Url:         null.StringFrom("http://localhost:9200"),
		CloudID:     null.NewString("", false),
		CACert:      null.NewString("", false),
		User:        null.NewString("", false),
		Password:    null.NewString("", false),
		FlushPeriod: types.NullDurationFrom(defaultFlushPeriod),
	}
}

// From here till the end of the file partial duplicates waiting for config refactor (k6 #883)

func (base Config) Apply(applied Config) Config {
	if applied.Url.Valid {
		base.Url = applied.Url
	}
	if applied.CloudID.Valid {
		base.CloudID = applied.CloudID
	}

	if applied.CACert.Valid {
		base.CACert = applied.CACert
	}

	if applied.User.Valid {
		base.User = applied.User
	}

	if applied.Password.Valid {
		base.Password = applied.Password
	}

	if applied.FlushPeriod.Valid {
		base.FlushPeriod = applied.FlushPeriod
	}

	return base
}

// ParseArg takes an arg string and converts it to a config
func ParseArg(arg string) (Config, error) {
	var c Config
	params, err := strvals.Parse(arg)
	if err != nil {
		return c, err
	}

	if v, ok := params["url"].(string); ok {
		c.Url = null.StringFrom(v)
	}

	if v, ok := params["cloud-id"].(string); ok {
		c.CloudID = null.StringFrom(v)
	}

	if v, ok := params["caCertFile"].(string); ok {
		c.CACert = null.StringFrom(v)
	}

	if v, ok := params["user"].(string); ok {
		c.User = null.StringFrom(v)
	}

	if v, ok := params["password"].(string); ok {
		c.Password = null.StringFrom(v)
	}

	if v, ok := params["flushPeriod"].(string); ok {
		if err := c.FlushPeriod.UnmarshalText([]byte(v)); err != nil {
			return c, err
		}
	}

	return c, nil
}

// GetConsolidatedConfig combines {default config values + JSON config +
// environment vars + arg config values}, and returns the final result.
func GetConsolidatedConfig(jsonRawConf json.RawMessage, env map[string]string, arg string) (Config, error) {
	result := NewConfig()
	if jsonRawConf != nil {
		jsonConf := Config{}
		if err := json.Unmarshal(jsonRawConf, &jsonConf); err != nil {
			return result, err
		}
		result = result.Apply(jsonConf)
	}

	// envconfig is not processing some undefined vars (at least duration) so apply them manually
	if flushPeriod, flushPeriodDefined := env["K6_ELASTICSEARCH_FLUSH_PERIOD"]; flushPeriodDefined {
		if err := result.FlushPeriod.UnmarshalText([]byte(flushPeriod)); err != nil {
			return result, err
		}
	}

	if url, urlDefined := env["K6_ELASTICSEARCH_URL"]; urlDefined {
		result.Url = null.StringFrom(url)
	}

	if cloudId, cloudIdDefined := env["K6_ELASTICSEARCH_CLOUD_ID"]; cloudIdDefined {
		result.CloudID = null.StringFrom(cloudId)
	}

	if ca, caDefined := env["K6_ELASTICSEARCH_CA_CERT_FILE"]; caDefined {
		result.CACert = null.StringFrom(ca)
	}

	if user, userDefined := env["K6_ELASTICSEARCH_USER"]; userDefined {
		result.User = null.StringFrom(user)
	}

	if password, passwordDefined := env["K6_ELASTICSEARCH_PASSWORD"]; passwordDefined {
		result.Password = null.StringFrom(password)
	}

	if arg != "" {
		argConf, err := ParseArg(arg)
		if err != nil {
			return result, err
		}

		result = result.Apply(argConf)
	}

	return result, nil
}
