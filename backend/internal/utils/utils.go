package utils

import (
	"context"
	"log"
	"strings"

	"github.com/htchan/UserService/backend/pkg/grpc"
	"github.com/htchan/WebHistory/internal/config"
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

func FindUserByToken(token string, conf *config.UserServiceConfig) string {
	if client == nil {
		client = grpc.NewClient(conf.Addr)
	}
	tokenPermission := grpc.NewAuthenticateParams(token, conf.Token, "")
	result, err := client.Authenticate(ctx, tokenPermission)
	if err != nil {
		log.Printf("fail to authenticate: %s", err)
		return ""
	}
	log.Println("authenticate result: ", result)
	return *result.Result
}
