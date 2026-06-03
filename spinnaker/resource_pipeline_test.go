package spinnaker

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestEditAndEncodePipeline(t *testing.T) {
	t.Run("strips top-level schema", func(t *testing.T) {
		pipeline := map[string]interface{}{
			"schema": "1",
			"stages": []interface{}{},
		}

		encoded, err := editAndEncodePipeline(pipeline)
		require.NoError(t, err)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(encoded), &result))
		require.NotContains(t, result, "schema")
	})

	t.Run("strips cron trigger id while preserving other fields", func(t *testing.T) {
		pipeline := map[string]interface{}{
			"triggers": []interface{}{
				map[string]interface{}{
					"type":    "docker",
					"enabled": false,
				},
				map[string]interface{}{
					"id":             "558f6f87-e0f8-402b-baac-ed895ec9cef4",
					"type":           "cron",
					"cronExpression": "0 0 18 ? * MON-FRI *",
					"enabled":        true,
				},
			},
		}

		encoded, err := editAndEncodePipeline(pipeline)
		require.NoError(t, err)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(encoded), &result))

		triggers := result["triggers"].([]interface{})
		require.Len(t, triggers, 2)

		dockerTrigger := triggers[0].(map[string]interface{})
		require.Equal(t, "docker", dockerTrigger["type"])
		require.NotContains(t, dockerTrigger, "id")

		cronTrigger := triggers[1].(map[string]interface{})
		require.Equal(t, "cron", cronTrigger["type"])
		require.Equal(t, "0 0 18 ? * MON-FRI *", cronTrigger["cronExpression"])
		require.NotContains(t, cronTrigger, "id")
	})

	t.Run("strips stage id while preserving other fields", func(t *testing.T) {
		pipeline := map[string]interface{}{
			"stages": []interface{}{
				map[string]interface{}{
					"id":   "stage-uuid",
					"type": "wait",
					"name": "Wait",
				},
			},
		}

		encoded, err := editAndEncodePipeline(pipeline)
		require.NoError(t, err)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(encoded), &result))

		stages := result["stages"].([]interface{})
		require.Len(t, stages, 1)

		stage := stages[0].(map[string]interface{})
		require.Equal(t, "wait", stage["type"])
		require.Equal(t, "Wait", stage["name"])
		require.NotContains(t, stage, "id")
	})
}

func TestPipelineDiffSuppressFunc(t *testing.T) {
	configPipeline := `{
		"schema": "1",
		"triggers": [
			{
				"type": "docker",
				"enabled": false
			},
			{
				"type": "cron",
				"cronExpression": "0 0 18 ? * MON-FRI *",
				"enabled": true
			}
		]
	}`

	statePipeline := `{
		"triggers": [
			{
				"type": "docker",
				"enabled": false
			},
			{
				"id": "558f6f87-e0f8-402b-baac-ed895ec9cef4",
				"type": "cron",
				"cronExpression": "0 0 18 ? * MON-FRI *",
				"enabled": true
			}
		]
	}`

	require.True(t, pipelineDiffSuppressFunc("pipeline", statePipeline, configPipeline, &schema.ResourceData{}))
}
