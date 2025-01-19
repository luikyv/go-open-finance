package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/oidc"
	"github.com/luikyv/go-open-finance/internal/user"
)

func openidProvider(userService user.Service, consentService consent.Service) (provider.Provider, error) {

	// Get the file path of the source file.
	_, filename, _, _ := runtime.Caller(0)
	sourceDir := filepath.Dir(filename)

	templatesDirPath := filepath.Join(sourceDir, "../../templates")
	// TODO: This will cause problems for the docker file.
	keysDir := filepath.Join(sourceDir, "../../keys")
	serverJWKS := privateJWKS(filepath.Join(keysDir, "server.jwks"))

	return provider.New(
		goidc.ProfileFAPI1,
		host,
		func(_ context.Context) (goidc.JSONWebKeySet, error) {
			return serverJWKS, nil
		},
		provider.WithPathPrefix(pathPrefixOIDC),
		provider.WithScopes(oidc.Scopes...),
		provider.WithTokenOptions(oidc.TokenOptionsFunc()),
		provider.WithAuthorizationCodeGrant(),
		provider.WithImplicitGrant(),
		provider.WithRefreshTokenGrant(oidc.ShoudIssueRefreshTokenFunc(), 600),
		provider.WithClientCredentialsGrant(),
		provider.WithTokenAuthnMethods(goidc.ClientAuthnPrivateKeyJWT),
		provider.WithPrivateKeyJWTSignatureAlgs(goidc.PS256),
		provider.WithMTLS(mtlsHost, oidc.ClientCertFunc()),
		provider.WithTLSCertTokenBindingRequired(),
		provider.WithPAR(60),
		provider.WithJAR(goidc.PS256),
		provider.WithJAREncryption(goidc.RSA_OAEP),
		provider.WithJARContentEncryptionAlgs(goidc.A256GCM),
		provider.WithJARM(goidc.PS256),
		provider.WithIssuerResponseParameter(),
		provider.WithPKCE(goidc.CodeChallengeMethodSHA256),
		provider.WithACRs(oidc.ACROpenBankingLOA2, oidc.ACROpenBankingLOA3),
		provider.WithUserInfoSignatureAlgs(goidc.PS256),
		provider.WithUserInfoEncryption(goidc.RSA_OAEP),
		provider.WithIDTokenSignatureAlgs(goidc.PS256),
		provider.WithIDTokenEncryption(goidc.RSA_OAEP),
		provider.WithStaticClient(client("client_one", keysDir)),
		provider.WithStaticClient(client("client_two", keysDir)),
		provider.WithHandleGrantFunc(oidc.HandleGrantFunc(consentService)),
		provider.WithPolicy(oidc.Policy(templatesDirPath, host+pathPrefixOIDC, userService, consentService)),
		provider.WithNotifyErrorFunc(oidc.LogErrorFunc()),
		provider.WithDCR(oidc.DCRFunc(oidc.Scopes), func(r *http.Request, s string) error {
			return nil
		}),
		provider.WithHTTPClientFunc(httpClientFunc()),
	)
}

func client(clientID string, keysDir string) *goidc.Client {
	var scopes []string
	for _, scope := range oidc.Scopes {
		scopes = append(scopes, scope.ID)
	}

	privateJWKS := privateJWKS(filepath.Join(keysDir, clientID+".jwks"))
	publicJWKS := goidc.JSONWebKeySet{Keys: []goidc.JSONWebKey{}}
	for _, jwk := range privateJWKS.Keys {
		publicJWKS.Keys = append(publicJWKS.Keys, jwk.Public())
	}
	rawPublicJWKS, _ := json.Marshal(publicJWKS)
	return &goidc.Client{
		ID: clientID,
		ClientMetaInfo: goidc.ClientMetaInfo{
			TokenAuthnMethod: goidc.ClientAuthnPrivateKeyJWT,
			ScopeIDs:         strings.Join(scopes, " "),
			RedirectURIs: []string{
				"https://localhost.emobix.co.uk:8443/test/a/mockbank/callback",
			},
			GrantTypes: []goidc.GrantType{
				goidc.GrantAuthorizationCode,
				goidc.GrantRefreshToken,
				goidc.GrantClientCredentials,
				goidc.GrantImplicit,
			},
			ResponseTypes: []goidc.ResponseType{
				goidc.ResponseTypeCode,
				goidc.ResponseTypeCodeAndIDToken,
			},
			PublicJWKS:           rawPublicJWKS,
			IDTokenKeyEncAlg:     goidc.RSA_OAEP,
			IDTokenContentEncAlg: goidc.A128CBC_HS256,
		},
	}
}

func privateJWKS(filePath string) goidc.JSONWebKeySet {
	absPath, _ := filepath.Abs(filePath)
	jwksFile, err := os.Open(absPath)
	if err != nil {
		log.Fatal(err)
	}
	defer jwksFile.Close()

	jwksBytes, err := io.ReadAll(jwksFile)
	if err != nil {
		log.Fatal(err)
	}

	var jwks goidc.JSONWebKeySet
	if err := json.Unmarshal(jwksBytes, &jwks); err != nil {
		log.Fatal(err)
	}

	return jwks
}

func httpClientFunc() goidc.HTTPClientFunc {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Renegotiation:      tls.RenegotiateOnceAsClient,
				InsecureSkipVerify: true,
			},
		},
	}

	return func(ctx context.Context) *http.Client {
		return client
	}
}
