Auth middleware for gin

[![codecov](https://codecov.io/gh/PxyUp/gin-auth-middleware/branch/master/graph/badge.svg)](https://codecov.io/gh/PxyUp/gin-auth-middleware)

# Use Auth middleware

```bash
go get github.com/PxyUp/gin-auth-middleware
```

```go
type UserFn func([]byte) (interface{}, error)

type User struct {
	Name string `json:"name"`
}

var (
	userFn = func([]byte) (interface{}, error) {
        user := &User{}
        err := json.Unmarshal(body, user)
        if err != nil {
            return nil, err
        }
        return user, nil
    } 
    
    mw = &gin_auth_middleware.MiddleWare{
        Host:        "http://localhost:8080"
        Path:        "/auth/me"
        Method:      http.MethodGet
        StatusCode:  http.StatusOk
        UserFn:      userFn
        Headers      map[string]string{}
        ProxyHeaders []string{"Authorization"}
    }
)


func CreateEngine() *gin.Engine {
	r := gin.New()
	r.Use(mw.Auth())
	return r
}

```

# Use GetUserFromContext
When user already authenticated you can user like that:

```go
func (c *gin.Context) {
	userD,err := gin_auth_middleware.GetUserFromContext(c)
	if err != nil {
		...
	}
	
	user := userD.(User) // cast type from user function
}
```