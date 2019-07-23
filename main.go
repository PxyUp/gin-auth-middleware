package gin_auth_middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

type MiddleWare struct {
	Host         string
	Path         string
	Method       string
	StatusCode   int
	UserFn       UserFn
	Headers      map[string]string
	ProxyHeaders []string
}

type UserFn func([]byte) (interface{}, error)

const (
	MaxIdleConnections int = 30
	RequestTimeout     int = 10
	USER_KEY               = "__gin_auth_middleware__"
)

var (
	httpClient             *http.Client
	CANT_PROCESS_USER_DATA = errors.New("CANT_PROCESS_USER_DATA")
	UNEXCEPTED_STATUS_CODE = errors.New("UNEXCEPTED_STATUS_CODE")
)

func init() {
	httpClient = createHTTPClient()
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

func (m *MiddleWare) Auth() gin.HandlerFunc {
	if m.UserFn == nil {
		panic("UserFn must set for processing user")
	}
	return func(context *gin.Context) {
		body, err := m.sendRequest(context)

		if err != nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, err := m.UserFn(body)

		if err != nil {
			context.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		context.Set(USER_KEY, user)
		context.Next()
	}
}

func GetUserFromContext(context *gin.Context) (interface{}, error) {
	userData, ok := context.Get(USER_KEY)
	if !ok {
		return nil, CANT_PROCESS_USER_DATA
	}
	return userData, nil
}

func (m *MiddleWare) sendRequest(c *gin.Context) ([]byte, error) {
	req, err := http.NewRequest(m.Method, m.Host+m.Path, nil)

	if err != nil {
		return nil, err
	}

	for _, k := range m.ProxyHeaders {
		req.Header.Add(k, c.GetHeader(k))
	}

	for k, v := range m.Headers {
		req.Header.Add(k, v)
	}

	res, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != m.StatusCode {
		return nil, UNEXCEPTED_STATUS_CODE
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
