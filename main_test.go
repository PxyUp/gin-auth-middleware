package gin_auth_middleware

import (
	"github.com/gin-gonic/gin"
	 "github.com/stretchr/testify/assert"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

var (
	mw = MiddleWare{
		Host:         "http://localhost:9333",
		Method:       "GET",
		Path:         "/auth/me",
		ProxyHeaders: []string{"Authorization"},
		StatusCode:   http.StatusOK,
		UserFn: func(bytes []byte) (i interface{}, e error) {
			return nil, nil
		},
	}
)

func init() {
	loginService := createLoginServerMock()
	loginListener, _ := net.Listen("tcp", ":9333")
	servLoginSv := &http.Server{Handler: loginService}
	go func() { log.Fatal(servLoginSv.Serve(loginListener)) }()
	testService := createTestServer()
	testListener, _ := net.Listen("tcp", ":9334")
	testSv := &http.Server{Handler: testService}
	go func() { log.Fatal(testSv.Serve(testListener)) }()
}

func TestMiddleWare_Auth(t *testing.T) {
	code := sendRequest(http.MethodGet, "http://localhost:9334/me", map[string]string{})
	assert.Equal(t, 401, code)
	codeSecond := sendRequest(http.MethodGet, "http://localhost:9334/me", map[string]string{
		"Authorization": "test",
	})
	assert.Equal(t, 200, codeSecond)
}

func createLoginServerMock() *gin.Engine {
	r := gin.New()
	r.GET("/auth/me", func(context *gin.Context) {
		value := context.Request.Header["Authorization"]
		if value[0] == "test" {
			context.JSON(http.StatusOK, gin.H{})
			context.Done()
			return
		}
		context.AbortWithStatus(401)
	})
	return r
}

func createTestServer() *gin.Engine {
	r := gin.New()
	r.Use(mw.Auth())
	r.GET("/me", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{})
	})
	return r
}

func sendRequest(method string, url string, headers map[string]string) int {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	req, _ := http.NewRequest(method, url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, _ := httpClient.Do(req)
	defer res.Body.Close()
	return res.StatusCode
}
