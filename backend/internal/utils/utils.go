package utils

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/htchan/UserService/backend/pkg/grpc"
)

func IsSubSet(s1 string, s2 string) bool {
	if len(s1) < len(s2) {
		return false
	}
	for _, char := range strings.Split(s2, "") {
		if !strings.Contains(s1, char) {
			return false
		}
	}
	return true
}

var client grpc.Client
var ctx context.Context = context.Background()

func FindUserByToken(token string) string {
	if client == nil {
		client = grpc.NewClient(os.Getenv("USER_SERVICE_ADDR"))
	}
	serviceToken := os.Getenv("SERVICE_TOKEN")
	tokenPermission := grpc.NewAuthenticateParams(token, serviceToken, "")
	result, err := client.Authenticate(ctx, tokenPermission)
	if err != nil {
		log.Printf("fail to authenticate: %s", err)
		return ""
	}
	log.Println("authenticate result: ", result)
	return *result.Result
}
