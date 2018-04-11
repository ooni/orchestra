package middleware

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellais/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	common "github.com/ooni/orchestra/common"
	"golang.org/x/crypto/bcrypt"
)

// This is taken from:
// https://github.com/appleboy/gin-jwt/blob/master/auth_jwt.go
// with some minor changes.

// OrchestraClaims are claims for the JWT token
type OrchestraClaims struct {
	Role string `json:"role"`
	User string `json:"user"`
	jwt.StandardClaims
}

// Account is the details of the account
type Account struct {
	Username string
	Role     string
}

// GinJWTMiddleware provides a Json-Web-Token authentication implementation. On failure, a 401 HTTP response
// is returned. On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userID").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type GinJWTMiddleware struct {
	// Realm name to display to the user. Required.
	Realm string

	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Secret key used for signing. Required.
	Key []byte

	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration

	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// Callback function that should perform the authentication of the user based on userID and
	// password. Must return true on success, false on failure. Required.
	// Option return user id, if so, user id will be stored in Claim Array.
	Authenticator func(userID string, password string, c *gin.Context) (Account, bool)

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via c.Get("JWT_PAYLOAD").

	// User can define own Unauthorized func.
	Unauthorized func(*gin.Context, int, string)

	// Set the identity handler function
	IdentityHandler func(*OrchestraClaims) Account

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName string

	// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	TimeFunc func() time.Time
}

// Authorizator structure
// Callback function that should perform the authorization of the authenticated user. Called
// only after an authentication success. Must return true on success, false on failure.
type Authorizator func(account Account, c *gin.Context) bool

// Login form structure.
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// MiddlewareInit initialize jwt configs.
func (mw *GinJWTMiddleware) MiddlewareInit() error {

	if mw.TokenLookup == "" {
		mw.TokenLookup = "header:Authorization"
	}

	if mw.SigningAlgorithm == "" {
		mw.SigningAlgorithm = "HS256"
	}

	if mw.Timeout == 0 {
		mw.Timeout = time.Hour
	}

	if mw.TimeFunc == nil {
		mw.TimeFunc = time.Now
	}

	mw.TokenHeadName = strings.TrimSpace(mw.TokenHeadName)
	if len(mw.TokenHeadName) == 0 {
		mw.TokenHeadName = "Bearer"
	}

	if mw.Unauthorized == nil {
		mw.Unauthorized = func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		}
	}

	if mw.IdentityHandler == nil {
		mw.IdentityHandler = func(claims *OrchestraClaims) Account {
			return Account{
				Username: claims.User,
				Role:     claims.Role,
			}
		}
	}

	if mw.Realm == "" {
		return errors.New("realm is required")
	}

	if mw.Key == nil {
		return errors.New("secret key is required")
	}

	return nil
}

// MiddlewareFunc makes GinJWTMiddleware implement the Middleware interface.
func (mw *GinJWTMiddleware) MiddlewareFunc(auth Authorizator) gin.HandlerFunc {
	if err := mw.MiddlewareInit(); err != nil {
		return func(c *gin.Context) {
			mw.unauthorized(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	return func(c *gin.Context) {
		mw.middlewareImpl(auth, c)
		return
	}
}

func (mw *GinJWTMiddleware) middlewareImpl(auth Authorizator, c *gin.Context) {
	token, err := mw.parseToken(c)

	// When the token is missing we still run the auth function to support the
	// NullAuthorizor It's very important that when you implement an Authorizor
	// you take this into account and that the default is that it returns false
	if err == ErrMissingToken {
		var defaultAccount = Account{
			Username: "",
			Role:     "unauthenticated",
		}
		if !auth(defaultAccount, c) {
			mw.unauthorized(c, http.StatusUnauthorized, err.Error())
			return
		}
		c.Next()
		return
	}

	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, err.Error())
		return
	}

	claims := token.Claims.(*OrchestraClaims)

	account := mw.IdentityHandler(claims)
	c.Set("JWT_PAYLOAD", claims)
	c.Set("userID", account.Username)

	if !auth(account, c) {
		mw.unauthorized(c, http.StatusForbidden, "You don't have permission to access.")
		return
	}

	c.Next()
}

// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GinJWTMiddleware) LoginHandler(c *gin.Context) {

	// Initial middleware default setting.
	mw.MiddlewareInit()

	var loginVals Login

	if c.BindJSON(&loginVals) != nil {
		mw.unauthorized(c, http.StatusBadRequest, "missing-username-password")
		return
	}

	if mw.Authenticator == nil {
		mw.unauthorized(c, http.StatusInternalServerError, "Missing define authenticator func")
		return
	}

	account, ok := mw.Authenticator(loginVals.Username, loginVals.Password, c)

	if !ok {
		mw.unauthorized(c, http.StatusUnauthorized, "wrong-username-password")
		return
	}

	// Create the token
	expire := mw.TimeFunc().Add(mw.Timeout)
	claims := OrchestraClaims{
		Role: account.Role,
		User: account.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
			IssuedAt:  mw.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod(mw.SigningAlgorithm), claims)

	tokenString, err := token.SignedString(mw.Key)

	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, "Create JWT Token faild")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  tokenString,
		"expire": expire.Format(time.RFC3339),
	})
}

// RefreshHandler can be used to refresh a token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the GinJWTMiddleware.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GinJWTMiddleware) RefreshHandler(c *gin.Context) {
	token, _ := mw.parseToken(c)
	claims := token.Claims.(*OrchestraClaims)

	origIat := int64(claims.StandardClaims.IssuedAt)

	if origIat < mw.TimeFunc().Add(-mw.MaxRefresh).Unix() {
		mw.unauthorized(c, http.StatusUnauthorized, "Token is expired.")
		return
	}

	// Create the token
	expire := mw.TimeFunc().Add(mw.Timeout)
	newClaims := OrchestraClaims{
		Role: claims.Role,
		User: claims.User,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
			IssuedAt:  origIat,
		},
	}
	newToken := jwt.NewWithClaims(jwt.GetSigningMethod(mw.SigningAlgorithm), newClaims)

	tokenString, err := newToken.SignedString(mw.Key)

	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, "Create JWT Token faild")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  tokenString,
		"expire": expire.Format(time.RFC3339),
	})
}

// ExtractClaims help to extract the JWT claims
func ExtractClaims(c *gin.Context) jwt.MapClaims {

	if _, exists := c.Get("JWT_PAYLOAD"); !exists {
		emptyClaims := make(jwt.MapClaims)
		return emptyClaims
	}

	jwtClaims, _ := c.Get("JWT_PAYLOAD")

	return jwtClaims.(jwt.MapClaims)
}

// TokenGenerator handler that clients can use to get a jwt token.
func (mw *GinJWTMiddleware) TokenGenerator(userID string, role string) string {
	// Create the token
	claims := OrchestraClaims{
		Role: userID,
		User: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: mw.TimeFunc().Add(mw.Timeout).Unix(),
			IssuedAt:  mw.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod(mw.SigningAlgorithm), claims)

	tokenString, _ := token.SignedString(mw.Key)

	return tokenString
}

// ErrMissingToken when the auth token is missing in the headers, query paramters or cookies
var ErrMissingToken = errors.New("JWT Token is missing")

func (mw *GinJWTMiddleware) jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)

	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == mw.TokenHeadName) {
		return "", errors.New("invalid auth header")
	}

	return parts[1], nil
}

func (mw *GinJWTMiddleware) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)

	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

func (mw *GinJWTMiddleware) jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)

	if cookie == "" {
		return "", ErrMissingToken
	}

	return cookie, nil
}

func (mw *GinJWTMiddleware) parseToken(c *gin.Context) (*jwt.Token, error) {
	var token string
	var err error

	parts := strings.Split(mw.TokenLookup, ":")
	switch parts[0] {
	case "header":
		token, err = mw.jwtFromHeader(c, parts[1])
	case "query":
		token, err = mw.jwtFromQuery(c, parts[1])
	case "cookie":
		token, err = mw.jwtFromCookie(c, parts[1])
	}

	if err != nil {
		return nil, err
	}

	return jwt.ParseWithClaims(token, &OrchestraClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if jwt.GetSigningMethod(mw.SigningAlgorithm) != token.Method {
				return nil, errors.New("invalid signing algorithm")
			}

			return mw.Key, nil
		})
}

func (mw *GinJWTMiddleware) unauthorized(c *gin.Context, code int, message string) {

	if mw.Realm == "" {
		mw.Realm = "gin jwt"
	}

	c.Header("WWW-Authenticate", "JWT realm="+mw.Realm)
	c.Abort()

	mw.Unauthorized(c, code, message)

	return
}

// InitAuthMiddleware is called to initialise the authentication middleware
func InitAuthMiddleware(db *sqlx.DB) (*GinJWTMiddleware, error) {
	return &GinJWTMiddleware{
		Realm:      "Proteus Realm", // XXX rename this to Orchestra
		Key:        []byte(viper.GetString("auth.jwt-token")),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (Account, bool) {
			var (
				passwordHash string
				account      Account
			)
			account.Username = userId
			if account.Username == "admin" {
				if viper.IsSet("auth.admin-password") == false {
					return account, false
				}
				adminPassword := []byte(viper.GetString("auth.admin-password"))
				if subtle.ConstantTimeCompare([]byte(password), adminPassword) == 1 {
					account.Role = "admin"
					return account, true
				}
				return account, false
			}
			// XXX set the last_login value
			query := fmt.Sprintf(`SELECT
							password_hash, role
							FROM %s WHERE username = $1`,
				pq.QuoteIdentifier(common.AccountsTable))
			err := db.QueryRow(query, userId).Scan(
				&passwordHash,
				&account.Role)
			if err != nil {
				if err == sql.ErrNoRows {
					return account, false
				}
				return account, false
			}
			err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
			if err != nil {
				return account, false
			}
			return account, true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":  code,
				"error": message,
			})
		},
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}, nil
}

// AdminAuthorizor is used to protect routes that are allowed only by administrator accounts
func AdminAuthorizor(account Account, c *gin.Context) bool {
	if account.Role == "admin" {
		return true
	}
	return false
}

// DeviceAuthorizor is used to protect routes that are allowed only by authenticated devices
func DeviceAuthorizor(account Account, c *gin.Context) bool {
	if account.Role == "device" {
		return true
	}
	return false
}

// NullAuthorizor is used for routes where authentication is optional and it
// returns always true
func NullAuthorizor(account Account, c *gin.Context) bool {
	return true
}
