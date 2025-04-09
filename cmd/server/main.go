package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/luikyv/go-open-finance/internal/account"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/creditcard"
	"github.com/luikyv/go-open-finance/internal/customer"
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
	creditCardStorage := creditcard.NewStorage()

	// Services.
	userService := user.NewService(userStorage)
	consentService := consent.NewService(consentStorage)
	resourceService := resource.NewService(consentService)
	customerService := customer.NewService(customerStorage)
	accountService := account.NewService(accountStorage, consentService)
	creditCardService := creditcard.NewService(creditCardStorage, consentService)

	// OpenID Provider.
	op, err := openidProvider(db, userService, consentService)
	if err != nil {
		log.Fatal(err)
	}

	// API Routers.
	consentAPIRouterV3 := consent.NewAPIRouterV3(mtlsHost, consentService, op)
	resourceAPIRouterV3 := resource.NewAPIRouterV3(mtlsHost, resourceService, consentService, op)
	customerAPIRouterV2 := customer.NewAPIRouterV2(mtlsHost, customerService, consentService, op)
	accountAPIRouterV2 := account.NewAPIRouterV2(mtlsHost, accountService, consentService, op)
	creditCardAPIRouterV2 := creditcard.NewAPIRouterV2(mtlsHost, creditCardService, consentService, op)

	// Server.
	mux := http.NewServeMux()

	mux.Handle(pathPrefixOIDC+"/", op.Handler())
	consentAPIRouterV3.Register(mux)
	resourceAPIRouterV3.Register(mux)
	customerAPIRouterV2.Register(mux)
	accountAPIRouterV2.Register(mux)
	creditCardAPIRouterV2.Register(mux)

	// Run.
	_ = loadMocks(userService, customerService, accountService, creditCardService)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func dbConnection() (*mongo.Database, error) {
	ctx := context.Background()

	opts := options.Client()
	opts = opts.ApplyURI(dbStringCon)
	opts = opts.SetBSONOptions(&options.BSONOptions{
		UseJSONStructTags: true,
		NilMapAsEmpty:     true,
		NilSliceAsEmpty:   true,
	})
	conn, err := mongo.Connect(ctx, opts)
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
