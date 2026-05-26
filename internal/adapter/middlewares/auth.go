package middlewares

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	accountUseCase "github.com/dev-au/CodeStream/internal/usecase/account"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	uc accountUseCase.UseCase
}

type credentials struct {
	ID       int
	Password string
}

func decodeAuthToken(token string) (*credentials, error) {
	if decodedBytes, err := base64.StdEncoding.DecodeString(token); err != nil {
		return nil, err
	} else {
		convertedByteToString := string(decodedBytes)
		splittedCredentials := strings.Split(convertedByteToString, ":")
		id, err := strconv.Atoi(splittedCredentials[0])
		if err != nil {
			return nil, err
		}
		return &credentials{
			ID:       id,
			Password: splittedCredentials[1],
		}, nil
	}
}

func (a *AuthMiddleware) SetCredentials(ctx *gin.Context) {
	abortUnauthorized := func() {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	authToken := ctx.GetHeader("Authorization")
	if len(authToken) < 5 {
		abortUnauthorized()
		return
	}

	crds, err := decodeAuthToken(authToken[5:])
	if err != nil {
		abortUnauthorized()
		return
	}

	user, err := a.uc.GetUser(ctx, crds.ID)
	if err != nil {
		abortUnauthorized()
		return
	}

	ctx.Set("user", user)
	ctx.Next()
}
