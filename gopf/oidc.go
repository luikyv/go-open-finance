package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/go-jose/go-jose/v4"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/goidcp"
	"github.com/luikyv/go-opf/gopf/constants"
	"go.mongodb.org/mongo-driver/mongo"
)

func Provider(db *mongo.Database) *goidcp.Provider {

	ps256ServerKeyID := "ps256_key"
	scopes := []goidc.Scope{goidc.ScopeOpenID, goidc.ScopeOffilineAccess, goidc.ScopeEmail,
		constants.ScopeConsent, constants.ScopeConsents, constants.ScopeCustomers, constants.ScopeAccounts,
		constants.ScopeCreditCardAccounts, constants.ScopeLoans, constants.ScopeFinancings,
		constants.ScopeUnarrangedAccountsOverdraft, constants.ScopeInvoiceFinancings, constants.ScopeBankFixedIncomes,
		constants.ScopeCreditFixedIncomes, constants.ScopeVariableIncomes, constants.ScopeTreasureTitles,
		constants.ScopeFunds, constants.ScopeExchanges, constants.ScopeResources}

	provider := goidcp.New(
		constants.Host,
		goidcp.NewMongoDBClientManager(db),
		goidcp.NewMongoDBAuthnSessionManager(db),
		goidcp.NewMongoDBGrantSessionManager(db),
		privateJWKS("../keys/server_jwks.json"),
		ps256ServerKeyID,
		ps256ServerKeyID,
	)
	provider.SetPathPrefix(constants.APIPrefixOIDC)
	provider.SetProfileFAPI2()
	provider.EnableMTLS(constants.MTLSHost)
	provider.EnableTLSBoundTokens()
	provider.RequirePushedAuthorizationRequests(60)
	provider.EnableJWTSecuredAuthorizationRequests(600, jose.PS256)
	provider.EnableJWTSecuredAuthorizationResponseMode(600, ps256ServerKeyID)
	provider.EnablePrivateKeyJWTClientAuthn(600, jose.PS256)
	provider.EnableIssuerResponseParameter()
	provider.EnableClaimsParameter()
	provider.EnableDemonstrationProofOfPossesion(600, jose.PS256, jose.ES256)
	provider.RequireProofKeyForCodeExchange(goidc.CodeChallengeMethodSHA256)
	provider.EnableRefreshTokenGrantType(6000, false)
	provider.SetScopes(scopes...)
	provider.SetSupportedUserClaims(
		goidc.ClaimEmail,
		goidc.ClaimEmailVerified,
	)
	provider.SetSupportedAuthenticationContextReferences(
		goidc.ACRMaceIncommonIAPSilver,
		constants.ACROpenBankingLOA2,
		constants.ACROpenBankingLOA3,
	)
	provider.EnableDynamicClientRegistration(nil, true)
	provider.SetTokenOptions(func(client *goidc.Client, scopes string) (goidc.TokenOptions, error) {
		return goidc.NewJWTTokenOptions(ps256ServerKeyID, 600), nil
	})
	provider.EnableUserInfoEncryption([]jose.KeyAlgorithm{jose.RSA_OAEP_256}, []jose.ContentEncryption{jose.A128CBC_HS256})

	// Create Client Mocks.
	clientOnePrivateJWKS := privateJWKS("../keys/client_one_jwks.json")
	clientOnePublicJWKS := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{}}
	for _, jwk := range clientOnePrivateJWKS.Keys {
		clientOnePublicJWKS.Keys = append(clientOnePublicJWKS.Keys, jwk.Public())
	}
	rawClientOnePublicJWKS, _ := json.Marshal(clientOnePublicJWKS)
	provider.AddClient(&goidc.Client{
		ID: "client_one",
		ClientMetaInfo: goidc.ClientMetaInfo{
			AuthnMethod: goidc.ClientAuthnPrivateKeyJWT,
			Scopes:      goidc.Scopes(scopes).String(),
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
			PublicJWKS:                        rawClientOnePublicJWKS,
			IDTokenKeyEncryptionAlgorithm:     jose.RSA_OAEP,
			IDTokenContentEncryptionAlgorithm: jose.A128CBC_HS256,
		},
	})

	clientTwoPrivateJWKS := privateJWKS("../keys/client_two_jwks.json")
	clientTwoPublicJWKS := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{}}
	for _, jwk := range clientTwoPrivateJWKS.Keys {
		clientTwoPublicJWKS.Keys = append(clientTwoPublicJWKS.Keys, jwk.Public())
	}
	rawClientTwoPublicJWKS, _ := json.Marshal(clientTwoPublicJWKS)
	provider.AddClient(&goidc.Client{
		ID: "client_two",
		ClientMetaInfo: goidc.ClientMetaInfo{
			AuthnMethod: goidc.ClientAuthnPrivateKeyJWT,
			Scopes:      goidc.Scopes(scopes).String(),
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
			PublicJWKS: rawClientTwoPublicJWKS,
		},
	})

	return provider
}

func privateJWKS(filePath string) jose.JSONWebKeySet {
	absPath, _ := filepath.Abs(filePath)
	jwksFile, err := os.Open(absPath)
	if err != nil {
		panic(err.Error())
	}
	defer jwksFile.Close()

	jwksBytes, err := io.ReadAll(jwksFile)
	if err != nil {
		panic(err.Error())
	}

	var jwks jose.JSONWebKeySet
	if err := json.Unmarshal(jwksBytes, &jwks); err != nil {
		panic(err.Error())
	}

	return jwks
}
