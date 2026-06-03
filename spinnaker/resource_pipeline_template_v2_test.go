package spinnaker

import (
	"encoding/json"
	"testing"

	"github.com/aegaxs/terraform-provider-spinnaker/spinnaker/api"
	"github.com/stretchr/testify/require"
)

func TestParsePipelineTemplateV2(t *testing.T) {
	t.Run("must be valid json", func(t *testing.T) {
		_, err := parsePipelineTemplateV2("{invalid")
		require.EqualError(t, err, `invalid template json: invalid character 'i' looking for beginning of object key string`)
	})

	t.Run("validates parsed template", func(t *testing.T) {
		_, err := parsePipelineTemplateV2(`{"variables":[{}]}`)
		require.Error(t, err)
		require.Regexp(t, `^invalid template: \d+ errors occurred:.*`, err.Error())
	})

	t.Run("parses valid template json", func(t *testing.T) {
		template, err := parsePipelineTemplateV2(`{"schema":"v2","pipeline":{},"metadata":{"name":"bar","description":"baz","scopes":["global"]}}`)
		require.NoError(t, err)

		expected := &api.PipelineTemplateV2{
			Schema: "v2",
			Metadata: api.PipelineTemplateV2Metadata{
				Name:        "bar",
				Description: "baz",
				Scopes:      []string{"global"},
			},
			Pipeline: map[string]interface{}{},
		}

		require.Equal(t, expected, template)
	})

	t.Run("validates that id field is not set", func(t *testing.T) {
		_, err := parsePipelineTemplateV2(`{"schema":"v2","id":"foo","pipeline":{},"metadata":{"name":"bar","description":"baz","scopes":["global"]}}`)
		require.Error(t, err)
		require.Regexp(t, `.*field 'id' must not be set.*`, err.Error())
	})

	t.Run("parses variable default field", func(t *testing.T) {
		template, err := parsePipelineTemplateV2(`{
			"schema":"v2",
			"pipeline":{},
			"metadata":{"name":"bar","description":"baz","scopes":["global"]},
			"variables":[{"name":"network_mode","type":"string","default":"bridge"}]
		}`)
		require.NoError(t, err)
		require.Len(t, template.Variables, 1)
		require.Equal(t, "bridge", template.Variables[0].Default)
	})
}

func TestNormalizeTemplateV2JSON(t *testing.T) {
	t.Run("coalesces default into defaultValue", func(t *testing.T) {
		raw := `{
			"schema":"v2",
			"variables":[{"name":"network_mode","type":"string","default":"bridge"}]
		}`

		normalized, err := normalizeTemplateV2JSON(raw)
		require.NoError(t, err)

		var doc map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(normalized), &doc))

		vars := doc["variables"].([]interface{})
		variable := vars[0].(map[string]interface{})
		require.Equal(t, "bridge", variable["defaultValue"])
		require.NotContains(t, variable, "default")
	})

	t.Run("leaves defaultValue-only variables unchanged", func(t *testing.T) {
		raw := `{
			"schema":"v2",
			"variables":[{"name":"autodeploy_tag","type":"string","defaultValue":"develop-latest"}]
		}`

		normalized, err := normalizeTemplateV2JSON(raw)
		require.NoError(t, err)

		var doc map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(normalized), &doc))

		variable := doc["variables"].([]interface{})[0].(map[string]interface{})
		require.Equal(t, "develop-latest", variable["defaultValue"])
		require.NotContains(t, variable, "default")
	})

	t.Run("strips top-level id", func(t *testing.T) {
		raw := `{"schema":"v2","id":"backofficeservices","variables":[]}`

		normalized, err := normalizeTemplateV2JSON(raw)
		require.NoError(t, err)

		var doc map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(normalized), &doc))
		require.NotContains(t, doc, "id")
	})
}

func TestPipelineTemplateV2DiffSuppressFunc(t *testing.T) {
	configTemplate := `{
		"schema":"v2",
		"variables":[
			{"name":"network_mode","type":"string","default":"bridge"}
		]
	}`

	stateTemplate := `{
		"schema":"v2",
		"variables":[
			{"name":"network_mode","type":"string","defaultValue":"bridge"}
		]
	}`

	require.True(t, pipelineTemplateV2DiffSuppressFunc("template", stateTemplate, configTemplate, nil))

	replicasConfig := `{
		"schema":"v2",
		"variables":[
			{"name":"replicas","type":"int","default":1}
		]
	}`

	replicasState := `{
		"schema":"v2",
		"variables":[
			{"name":"replicas","type":"int","defaultValue":1}
		]
	}`

	require.True(t, pipelineTemplateV2DiffSuppressFunc("template", replicasState, replicasConfig, nil))
}
