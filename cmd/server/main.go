package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-open-finance/internal/account"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/api/middleware"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/customer"
	"github.com/luikyv/go-open-finance/internal/oidc"
	"github.com/luikyv/go-open-finance/internal/resource"
	"github.com/luikyv/go-open-finance/internal/user"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	host           = getEnv("MOCKBANK_HOST", "https://mockbank.local")
	mtlsHost       = getEnv("MOCKBANK_MTLS_HOST", "https://matls-mockbank.local")
	port           = getEnv("MOCKBANK_PORT", "80")
	dbSchema       = getEnv("MOCKBANK_DB_SCHEMA", "mockbank")
	dbStringCon    = getEnv("MOCKBANK_DB_CONNECTION", "mongodb://localhost:27017/mockbank")
	pathPrefixOIDC = "/auth"
)

func main() {
	// Logging.
	slog.SetDefault(logger())

	// Database.
	db, err := dbConnection()
	if err != nil {
		log.Fatal(err)
	}

	// Storage.
	userStorage := user.NewStorage()
	consentStorage := consent.NewStorage(db)
	customerStorage := customer.NewStorage()
	accountStorage := account.NewStorage()

	// Services.
	userService := user.NewService(userStorage)
	consentService := consent.NewService(consentStorage)
	resourceService := resource.NewService(consentService)
	customerService := customer.NewService(customerStorage)
	accountService := account.NewService(accountStorage, consentService)

	// API Handlers.
	customerAPIHandlerV2 := customer.NewAPIHandlerV2(customerService)
	consentAPIHandlerV3 := consent.NewAPIHandlerV3(mtlsHost, consentService)
	resourceAPIHandlerV3 := resource.NewAPIHandlerV3(resourceService)
	accountAPIHandlerV2 := account.NewAPIHandlerV2(accountService)

	// Server.
	mux := http.NewServeMux()

	op, err := openidProvider(db, userService, consentService)
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle(pathPrefixOIDC+"/", op.Handler())

	opfMux := http.NewServeMux()

	// TODO: Find an easier way to set this.
	consentMux := http.NewServeMux()
	consentMux.Handle("POST /open-banking/consents/v3/consents", middleware.AuthScopes(consentAPIHandlerV3.CreateHandler(), op, oidc.ScopeConsents))
	consentMux.Handle("GET /open-banking/consents/v3/consents/{id}", middleware.AuthScopes(consentAPIHandlerV3.GetHandler(), op, oidc.ScopeConsents))
	consentMux.Handle("DELETE /open-banking/consents/v3/consents/{id}", middleware.AuthScopes(consentAPIHandlerV3.DeleteHandler(), op, oidc.ScopeConsents))
	consentMux.Handle("POST /open-banking/consents/v3/consents/{id}/extends", middleware.AuthScopes(consentAPIHandlerV3.ExtendHandler(), op, goidc.ScopeOpenID, oidc.ScopeConsentID))
	consentMux.Handle("GET /open-banking/consents/v3/consents/{id}/extensions", middleware.AuthScopes(consentAPIHandlerV3.GetExtensionsHandler(), op, oidc.ScopeConsents))
	opfMux.Handle("/open-banking/consents/", consentMux)

	customerMux := http.NewServeMux()
	customerMux.Handle("GET /open-banking/customers/v2/personal/identifications",
		middleware.AuthPermissions(customerAPIHandlerV2.GetPersonalIdentificationsHandler(), consentService, consent.PermissionCustomersPersonalIdentificationsRead))
	customerMux.Handle("GET /open-banking/customers/v2/personal/qualifications",
		middleware.AuthPermissions(customerAPIHandlerV2.GetPersonalQualificationsHandler(), consentService, consent.PermissionCustomersPersonalAdittionalInfoRead))
	customerMux.Handle("GET /open-banking/customers/v2/personal/financial-relations",
		middleware.AuthPermissions(customerAPIHandlerV2.GetPersonalFinancialRelationsHandler(), consentService, consent.PermissionCustomersPersonalAdittionalInfoRead))
	opfMux.Handle("/open-banking/customers/", middleware.AuthScopes(customerMux, op, goidc.ScopeOpenID, oidc.ScopeConsentID))

	opfMux.Handle("GET /open-banking/resources/v3/resources",
		middleware.AuthScopes(middleware.AuthPermissions(resourceAPIHandlerV3.GetHandler(), consentService, consent.PermissionResourcesRead), op, oidc.ScopeConsentID))

	accountMux := http.NewServeMux()
	accountMux.Handle("GET /open-banking/accounts/v2/accounts",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountsHandler(), consentService, consent.PermissionAccountsRead))
	accountMux.Handle("GET /open-banking/accounts/v2/accounts/{id}",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountHandler(), consentService, consent.PermissionAccountsRead))
	accountMux.Handle("GET /open-banking/accounts/v2/accounts/{id}/balances",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountBalancesHandler(), consentService, consent.PermissionAccountsBalanceRead))
	accountMux.Handle("GET /open-banking/accounts/v2/accounts/{id}/transactions",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountTransactionsHandler(), consentService, consent.PermissionAccountsTransactionsRead))
	accountMux.Handle("GET /open-banking/accounts/v2/accounts/{id}/transactions-current",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountTransactionsHandler(), consentService, consent.PermissionAccountsTransactionsRead))
	accountMux.Handle("GET /open-banking/accounts/v2/accounts/{id}/overdraft-limits",
		middleware.AuthPermissions(accountAPIHandlerV2.GetAccountOverdraftLimitsHandler(), consentService, consent.PermissionAccountsOverdraftLimitsRead))
	opfMux.Handle("/open-banking/accounts/", middleware.AuthScopes(accountMux, op, goidc.ScopeOpenID, oidc.ScopeConsentID))

	mux.Handle("/open-banking/", middleware.Meta(middleware.FAPIID(opfMux), mtlsHost))

	// Run.
	_ = loadMocks(userService, customerService, accountService)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func dbConnection() (*mongo.Database, error) {
	ctx := context.Background()

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(dbStringCon).SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags: true,
		NilMapAsEmpty:     true,
		NilSliceAsEmpty:   true,
	}))
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return conn.Database(dbSchema), nil
}

// getEnv retrieves an environment variable or returns a fallback value if not found
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func logger() *slog.Logger {
	return slog.New(&logCtxHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			// Make sure time is logged in UTC.
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					utcTime := time.Now()
					return slog.Attr{Key: slog.TimeKey, Value: slog.TimeValue(utcTime)}
				}
				return a
			},
		}),
	})
}

type logCtxHandler struct {
	slog.Handler
}

func (h *logCtxHandler) Handle(ctx context.Context, r slog.Record) error {
	if clientID, ok := ctx.Value(api.CtxKeyClientID).(string); ok {
		r.AddAttrs(slog.String("client_id", clientID))
	}
	if interactionID, ok := ctx.Value(api.CtxKeyInteractionID).(string); ok {
		r.AddAttrs(slog.String("interaction_id", interactionID))
	}
	return h.Handler.Handle(ctx, r)
}
