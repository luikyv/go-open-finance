package api

type ContextKey string

const (
	CtxKeyClientID      ContextKey = "client_id"
	CtxKeySubject       ContextKey = "subject"
	CtxKeyConsentID     ContextKey = "consent_id"
	CtxKeyInteractionID ContextKey = "interaction_id"
	CtxKeyRequestURL    ContextKey = "request_url"
)
