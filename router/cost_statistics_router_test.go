package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCostStatisticsLogsRouteRequiresAdminAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetApiRouter(engine)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/cost-statistics/logs", nil)
	engine.ServeHTTP(recorder, request)

	// A missing route would return 404. The registered route reaches AdminAuth
	// and rejects an unauthenticated request before the controller is called.
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
