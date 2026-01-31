package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSObject_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"attributes": {"type": "Account", "url": "/services/data/v62.0/sobjects/Account/001xx"},
		"Id": "001xx000003ABCDEF",
		"Name": "Test Account",
		"AnnualRevenue": 1000000.50,
		"Active__c": true,
		"NumberOfEmployees": 100
	}`

	var obj SObject
	err := json.Unmarshal([]byte(jsonData), &obj)
	require.NoError(t, err)

	assert.Equal(t, "Account", obj.Attributes.Type)
	assert.Equal(t, "/services/data/v62.0/sobjects/Account/001xx", obj.Attributes.URL)
	assert.Equal(t, "001xx000003ABCDEF", obj.ID)
	assert.Equal(t, "Test Account", obj.Fields["Name"])
	assert.Equal(t, 1000000.50, obj.Fields["AnnualRevenue"])
	assert.Equal(t, true, obj.Fields["Active__c"])
	assert.Equal(t, float64(100), obj.Fields["NumberOfEmployees"])
}

func TestSObject_MarshalJSON(t *testing.T) {
	obj := SObject{
		Attributes: SObjectAttributes{Type: "Account"},
		ID:         "001xx000003ABCDEF",
		Fields: map[string]interface{}{
			"Name":          "Test Account",
			"AnnualRevenue": 1000000.50,
		},
	}

	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "001xx000003ABCDEF", result["Id"])
	assert.Equal(t, "Test Account", result["Name"])
	assert.Equal(t, 1000000.50, result["AnnualRevenue"])
}

func TestSObject_GetString(t *testing.T) {
	obj := SObject{
		Fields: map[string]interface{}{
			"Name":   "Test Account",
			"Number": 123,
		},
	}

	assert.Equal(t, "Test Account", obj.GetString("Name"))
	assert.Equal(t, "", obj.GetString("Number"))    // wrong type
	assert.Equal(t, "", obj.GetString("NotExists")) // doesn't exist
}

func TestSObject_GetBool(t *testing.T) {
	obj := SObject{
		Fields: map[string]interface{}{
			"Active":   true,
			"Inactive": false,
			"Name":     "Test",
		},
	}

	assert.True(t, obj.GetBool("Active"))
	assert.False(t, obj.GetBool("Inactive"))
	assert.False(t, obj.GetBool("Name"))      // wrong type
	assert.False(t, obj.GetBool("NotExists")) // doesn't exist
}

func TestSObject_GetFloat(t *testing.T) {
	obj := SObject{
		Fields: map[string]interface{}{
			"Revenue": 1000000.50,
			"Count":   float64(100),
			"Name":    "Test",
		},
	}

	assert.Equal(t, 1000000.50, obj.GetFloat("Revenue"))
	assert.Equal(t, float64(100), obj.GetFloat("Count"))
	assert.Equal(t, float64(0), obj.GetFloat("Name"))      // wrong type
	assert.Equal(t, float64(0), obj.GetFloat("NotExists")) // doesn't exist
}

func TestSObject_GetInt(t *testing.T) {
	obj := SObject{
		Fields: map[string]interface{}{
			"Count":   float64(100),
			"Revenue": 1000000.75,
		},
	}

	assert.Equal(t, 100, obj.GetInt("Count"))
	assert.Equal(t, 1000000, obj.GetInt("Revenue")) // truncates decimal
}

func TestSObject_GetTime(t *testing.T) {
	obj := SObject{
		Fields: map[string]interface{}{
			"CreatedDate": "2024-01-15T10:30:00.000+0000",
			"Name":        "Test",
		},
	}

	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.FixedZone("", 0))
	assert.Equal(t, expected, obj.GetTime("CreatedDate"))
	assert.True(t, obj.GetTime("Name").IsZero())      // wrong type
	assert.True(t, obj.GetTime("NotExists").IsZero()) // doesn't exist
}

func TestSObject_NilFields(t *testing.T) {
	obj := SObject{}

	// Should not panic with nil Fields map
	assert.Equal(t, "", obj.GetString("Name"))
	assert.False(t, obj.GetBool("Active"))
	assert.Equal(t, float64(0), obj.GetFloat("Revenue"))
	assert.Equal(t, 0, obj.GetInt("Count"))
}

func TestQueryResult_Unmarshal(t *testing.T) {
	jsonData := `{
		"totalSize": 2,
		"done": true,
		"records": [
			{"attributes": {"type": "Account"}, "Id": "001", "Name": "Account 1"},
			{"attributes": {"type": "Account"}, "Id": "002", "Name": "Account 2"}
		]
	}`

	var result QueryResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.Equal(t, 2, result.TotalSize)
	assert.True(t, result.Done)
	assert.Len(t, result.Records, 2)
	assert.Equal(t, "001", result.Records[0].ID)
	assert.Equal(t, "Account 1", result.Records[0].GetString("Name"))
}

func TestQueryResult_WithNextRecords(t *testing.T) {
	jsonData := `{
		"totalSize": 5000,
		"done": false,
		"nextRecordsUrl": "/services/data/v62.0/query/01gxx0000000001-2000",
		"records": []
	}`

	var result QueryResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.Equal(t, 5000, result.TotalSize)
	assert.False(t, result.Done)
	assert.Equal(t, "/services/data/v62.0/query/01gxx0000000001-2000", result.NextRecordsURL)
}

func TestRecordResult_Success(t *testing.T) {
	jsonData := `{
		"id": "001xx000003ABCDEF",
		"success": true,
		"errors": []
	}`

	var result RecordResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.Equal(t, "001xx000003ABCDEF", result.ID)
	assert.True(t, result.Success)
	assert.Empty(t, result.Errors)
}

func TestRecordResult_WithErrors(t *testing.T) {
	jsonData := `{
		"success": false,
		"errors": [
			{"statusCode": "REQUIRED_FIELD_MISSING", "message": "Required fields are missing: [Name]", "fields": ["Name"]}
		]
	}`

	var result RecordResult
	err := json.Unmarshal([]byte(jsonData), &result)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "REQUIRED_FIELD_MISSING", result.Errors[0].StatusCode)
	assert.Contains(t, result.Errors[0].Fields, "Name")
}

func TestSObjectDescribe(t *testing.T) {
	jsonData := `{
		"name": "Account",
		"label": "Account",
		"labelPlural": "Accounts",
		"keyPrefix": "001",
		"custom": false,
		"createable": true,
		"updateable": true,
		"deletable": true,
		"queryable": true,
		"searchable": true,
		"fields": [
			{
				"name": "Id",
				"label": "Account ID",
				"type": "id",
				"nillable": false,
				"createable": false,
				"updateable": false
			},
			{
				"name": "Name",
				"label": "Account Name",
				"type": "string",
				"length": 255,
				"nillable": false,
				"createable": true,
				"updateable": true
			}
		]
	}`

	var desc SObjectDescribe
	err := json.Unmarshal([]byte(jsonData), &desc)
	require.NoError(t, err)

	assert.Equal(t, "Account", desc.Name)
	assert.Equal(t, "Accounts", desc.LabelPlural)
	assert.Equal(t, "001", desc.KeyPrefix)
	assert.True(t, desc.Createable)
	assert.True(t, desc.Queryable)
	assert.Len(t, desc.Fields, 2)
	assert.Equal(t, "id", desc.Fields[0].Type)
	assert.Equal(t, 255, desc.Fields[1].Length)
}

func TestAPIVersion(t *testing.T) {
	jsonData := `{
		"label": "Spring '24",
		"url": "/services/data/v60.0",
		"version": "60.0"
	}`

	var version APIVersion
	err := json.Unmarshal([]byte(jsonData), &version)
	require.NoError(t, err)

	assert.Equal(t, "Spring '24", version.Label)
	assert.Equal(t, "/services/data/v60.0", version.URL)
	assert.Equal(t, "60.0", version.Version)
}

func TestLimits(t *testing.T) {
	jsonData := `{
		"DailyApiRequests": {"Max": 15000, "Remaining": 14500},
		"DailyBulkApiRequests": {"Max": 5000, "Remaining": 4999}
	}`

	var limits Limits
	err := json.Unmarshal([]byte(jsonData), &limits)
	require.NoError(t, err)

	assert.Equal(t, 15000, limits["DailyApiRequests"].Max)
	assert.Equal(t, 14500, limits["DailyApiRequests"].Remaining)
	assert.Equal(t, 5000, limits["DailyBulkApiRequests"].Max)
}

func TestCompositeRequest(t *testing.T) {
	req := CompositeRequest{
		AllOrNone: true,
		CompositeRequest: []CompositeSubrequest{
			{
				Method:      "POST",
				URL:         "/services/data/v62.0/sobjects/Account",
				ReferenceID: "newAccount",
				Body:        map[string]interface{}{"Name": "New Account"},
			},
			{
				Method:      "GET",
				URL:         "/services/data/v62.0/sobjects/Account/@{newAccount.id}",
				ReferenceID: "getAccount",
			},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result CompositeRequest
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.True(t, result.AllOrNone)
	assert.Len(t, result.CompositeRequest, 2)
	assert.Equal(t, "newAccount", result.CompositeRequest[0].ReferenceID)
}
