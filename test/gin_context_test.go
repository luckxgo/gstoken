package test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luckxgo/gstoken/web"
)

func TestGinContextDualContextSupport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	ginCtx := web.NewGinContext(c)

	testKey := "test_key"
	testValue := "test_value"

	ginCtx.Set(testKey, testValue)

	if value, exists := ginCtx.Get(testKey); !exists || value != testValue {
		t.Errorf("Failed to get value from Gin context: expected %v, got %v", testValue, value)
	}

	stdCtx := ginCtx.GetContext()
	if value := stdCtx.Value(testKey); value != testValue {
		t.Errorf("Failed to get value from standard context: expected %v, got %v", testValue, value)
	}
}

func TestGinContextWithBaseContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseCtx := context.WithValue(context.Background(), "base_key", "base_value")
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(baseCtx)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	ginCtx := web.NewGinContext(c)

	if value := ginCtx.GetContext().Value("base_key"); value != "base_value" {
		t.Errorf("Failed to get base context value: expected 'base_value', got %v", value)
	}

	ginCtx.Set("new_key", "new_value")

	stdCtx := ginCtx.GetContext()
	if value := stdCtx.Value("base_key"); value != "base_value" {
		t.Errorf("Lost base context value: expected 'base_value', got %v", value)
	}
	if value := stdCtx.Value("new_key"); value != "new_value" {
		t.Errorf("Failed to get new context value: expected 'new_value', got %v", value)
	}
}