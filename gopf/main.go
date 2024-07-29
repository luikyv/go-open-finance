package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-opf/gopf/api"
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/repositories"
	"github.com/luikyv/go-opf/gopf/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	db := dbConnection()

	// Repositories.
	userRepository := repositories.NewUser()
	consentRepository := repositories.NewConsent(db)

	// Services.
	userService := services.NewUser(userRepository)
	consentService := services.NewConsent(userService, consentRepository)

	// Provider.
	provider := Provider(db)
	provider.AddPolicy(goidc.NewPolicy(
		"policy",
		func(ctx goidc.Context, client *goidc.Client, session *goidc.AuthnSession) bool { return true },
		authenticationFunc(userService, consentService),
	))

	// APIs.
	server := gin.Default()

	openBankingRouter := server.Group(constants.APIPrefixOpenBanking)
	openBankingRouter.Use(api.FAPIHeaderMiddleware())
	api.RouteConsentsV3(openBankingRouter, provider, consentService)

	oidcRouter := server.Group(constants.APIPrefixOIDC)
	oidcRouter.Any("/*w", gin.WrapH(provider.Handler()))

	// Run.
	if err := server.Run(constants.Port); err != nil {
		panic(err)
	}
}

func dbConnection() *mongo.Database {
	options := options.Client().ApplyURI(constants.DatabaseStringConnection)
	conn, err := mongo.Connect(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return conn.Database(constants.DatabaseSchema)
}
