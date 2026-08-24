package google

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/domain"
	"golang.org/x/oauth2"
)

const issuerURL = "https://accounts.google.com"

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Provider struct {
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(ctx context.Context, cfg Config) (*Provider, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, err
	}

	return &Provider{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func (p *Provider) Name() domain.IdentityProvider {
	return domain.IdentityProviderGoogle
}

func (p *Provider) AuthorizationURL(state, nonce, codeChallenge string) string {
	return p.oauth.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (p *Provider) Authenticate(ctx context.Context, code, codeVerifier string) (auth.IdentityProfile, error) {
	token, err := p.oauth.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return auth.IdentityProfile{}, err
	}
	if token == nil {
		return auth.IdentityProfile{}, errors.New("provider response was empty")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return auth.IdentityProfile{}, errors.New("provider response did not include an ID token")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return auth.IdentityProfile{}, err
	}

	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return auth.IdentityProfile{}, err
	}

	return auth.IdentityProfile{
		Subject:       claims.Subject,
		Email:         claims.Email,
		DisplayName:   claims.Name,
		EmailVerified: claims.EmailVerified,
		Nonce:         claims.Nonce,
	}, nil
}
