package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type quotaDatesResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    []struct {
		ModelName string `json:"model_name"`
		TokenUsed int    `json:"token_used"`
		Count     int    `json:"count"`
		Quota     int    `json:"quota"`
	} `json:"data"`
}

func getQuotaDatesResponse(t *testing.T, recorder *httptest.ResponseRecorder) quotaDatesResponse {
	t.Helper()
	require.Equal(t, http.StatusOK, recorder.Code)
	var payload quotaDatesResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success, payload.Message)
	return payload
}

func TestGetAllQuotaDatesFiltersByTokenName(t *testing.T) {
	setupFlowControllerTestDB(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/data?start_timestamp=1000&end_timestamp=2000&token_name=primary", nil)

	GetAllQuotaDates(ctx)

	payload := getQuotaDatesResponse(t, recorder)
	require.Len(t, payload.Data, 1)
	require.Equal(t, "gpt-a", payload.Data[0].ModelName)
	require.Equal(t, 40, payload.Data[0].TokenUsed)
}

func TestGetAllQuotaDatesCombinesTokenNameAndUsername(t *testing.T) {
	setupFlowControllerTestDB(t)

	// "primary" belongs to alice and "backup" to bob, so the combined filter
	// must narrow down to the rows matching both dimensions.
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/data?start_timestamp=1000&end_timestamp=2000&username=bob&token_name=primary", nil)

	GetAllQuotaDates(ctx)

	payload := getQuotaDatesResponse(t, recorder)
	require.Empty(t, payload.Data)
}

func TestGetAllQuotaDatesTokenNameWithoutMatch(t *testing.T) {
	setupFlowControllerTestDB(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/data?start_timestamp=1000&end_timestamp=2000&token_name=nope", nil)

	GetAllQuotaDates(ctx)

	payload := getQuotaDatesResponse(t, recorder)
	require.Empty(t, payload.Data)
}

func TestGetUserQuotaDatesFiltersByOwnTokenName(t *testing.T) {
	setupFlowControllerTestDB(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/data/self?start_timestamp=1000&end_timestamp=2000&token_name=primary", nil)

	GetUserQuotaDates(ctx)

	payload := getQuotaDatesResponse(t, recorder)
	require.Len(t, payload.Data, 1)
	require.Equal(t, "gpt-a", payload.Data[0].ModelName)
	require.Equal(t, 40, payload.Data[0].TokenUsed)
}

func TestGetUserQuotaDatesRejectsAnotherUsersToken(t *testing.T) {
	setupFlowControllerTestDB(t)

	// "backup" belongs to user 2, so a self-service query from user 1 must not
	// match it even though both rows fall inside the requested time range.
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/data/self?start_timestamp=1000&end_timestamp=2000&token_name=backup", nil)

	GetUserQuotaDates(ctx)

	payload := getQuotaDatesResponse(t, recorder)
	require.Empty(t, payload.Data)
}
