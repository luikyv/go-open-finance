package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/api/middleware"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/oidc"
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

	// Services.
	userService := user.NewService(userStorage)
	consentService := consent.NewService(consentStorage)

	// API Handlers.
	userAPIHandlerV2 := user.NewAPIHandlerV2(userService)
	consentAPIHandlerV3 := consent.NewAPIHandlerV3(mtlsHost, consentService)

	// Server.
	mux := http.NewServeMux()

	op, err := openidProvider(db, userService, consentService)
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle(pathPrefixOIDC+"/", op.Handler())

	opfMux := http.NewServeMux()

	consentMux := http.NewServeMux()
	consentMux.Handle("POST /open-banking/consents/v3/consents", consentAPIHandlerV3.CreateHandler())
	consentMux.Handle("GET /open-banking/consents/v3/consents/{consent_id}", consentAPIHandlerV3.GetHandler())
	consentMux.Handle("DELETE /open-banking/consents/v3/consents/{consent_id}", consentAPIHandlerV3.DeleteHandler())
	opfMux.Handle("/open-banking/consents/", middleware.AuthScopes(consentMux, op, oidc.ScopeConsents))

	customersMux := http.NewServeMux()
	customersMux.Handle("GET /open-banking/customers/v2/personal/identifications",
		middleware.AuthPermissions(userAPIHandlerV2.GetPersonalIdentificationsHandler(), consentService, consent.PermissionCustomersPersonalIdentificationsRead))
	customersMux.Handle("GET /open-banking/customers/v2/personal/qualifications",
		middleware.AuthPermissions(userAPIHandlerV2.GetPersonalQualificationsHandler(), consentService, consent.PermissionCustomersPersonalAdittionalInfoRead))
	customersMux.Handle("GET /open-banking/customers/v2/personal/financial-relations",
		middleware.AuthPermissions(userAPIHandlerV2.GetPersonalFinancialRelationsHandler(), consentService, consent.PermissionCustomersPersonalAdittionalInfoRead))
	opfMux.Handle("/open-banking/customers/", middleware.AuthScopes(customersMux, op, oidc.ScopeConsentID))

	mux.Handle("/open-banking/", middleware.Meta(middleware.FAPIID(opfMux), mtlsHost))

	// Run.
	_ = loadMocks(userService)
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
