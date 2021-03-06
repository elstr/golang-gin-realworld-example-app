package users

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/recoilme/golang-gin-realworld-example-app/common"

	"github.com/gin-gonic/gin"
)

// Strips 'TOKEN ' prefix from token string
func stripBearerPrefixFromTokenString(tok string) (string, error) {
	// Should be a bearer token
	if len(tok) > 5 && strings.ToUpper(tok[0:6]) == "TOKEN " {
		return tok[6:], nil
	}
	return tok, nil
}

// Extract  token from Authorization header
// Uses PostExtractionFilter to strip "TOKEN " prefix from header
var AuthorizationHeaderExtractor = &request.PostExtractionFilter{
	request.HeaderExtractor{"Authorization"},
	stripBearerPrefixFromTokenString,
}

// Extractor for OAuth2 access tokens.  Looks in 'Authorization'
// header then 'access_token' argument for a token.
var MyAuth2Extractor = &request.MultiExtractor{
	AuthorizationHeaderExtractor,
	request.ArgumentExtractor{"access_token"},
}

// A helper to write user_id and user_model to the context
func UpdateContextUserModel(c *gin.Context, my_user_id uint32) {
	var myUserModel UserModel
	var err error
	if my_user_id != 0 {
		myUserModel, err = FindOneUser(&UserModel{ID: my_user_id})
		if err != nil {
			log.Println("UpdateContextUserModel:", err)
		} else {
			//TODO why not here?
			//c.Set("my_user_id", my_user_id)
			//c.Set("my_user_model", myUserModel)
		}
	}
	c.Set("my_user_id", my_user_id)
	c.Set("my_user_model", myUserModel)
	/*
		fmt.Println()
		fmt.Println("UpdateContextUserModel", my_user_id)
		fmt.Printf("%+v\n\n", myUserModel)
		fmt.Println()
	*/
}

// You can custom middlewares yourself as the doc: https://github.com/gin-gonic/gin#custom-middleware
//  r.Use(AuthMiddleware(true))
func AuthMiddleware(auto401 bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		UpdateContextUserModel(c, 0)
		token, err := request.ParseFromRequest(c.Request, MyAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			b := ([]byte(common.NBSecretPassword))
			return b, nil
		})
		if err != nil {
			if auto401 {
				c.AbortWithError(http.StatusUnauthorized, err)
			}
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			my_user_id := uint32(claims["id"].(float64))
			//fmt.Println("AuthMiddleware:", my_user_id, claims["id"])
			UpdateContextUserModel(c, my_user_id)
		}
	}
}
