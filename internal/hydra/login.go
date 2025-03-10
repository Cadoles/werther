/*
Copyright (c) JSC iCore.

This source code is licensed under the MIT license found in the
LICENSE file in the root directory of this source tree.
*/

package hydra

import (
	"github.com/pkg/errors"
)

// LoginReqDoer fetches information on the OAuth2 request and then accept or reject the requested authentication process.
type LoginReqDoer struct {
	hydraURL           string
	fakeTLSTermination bool
	rememberFor        int
	acr                string
	amr                []string
}

// NewLoginReqDoer creates a LoginRequest.
func NewLoginReqDoer(hydraURL string, fakeTLSTermination bool, rememberFor int, acr string, amr []string) *LoginReqDoer {
	return &LoginReqDoer{hydraURL: hydraURL, fakeTLSTermination: fakeTLSTermination, rememberFor: rememberFor, acr: acr, amr: amr}
}

// InitiateRequest fetches information on the OAuth2 request.
func (lrd *LoginReqDoer) InitiateRequest(challenge string) (*ReqInfo, error) {
	ri, err := initiateRequest(login, lrd.hydraURL, lrd.fakeTLSTermination, challenge)
	return ri, errors.Wrap(err, "failed to initiate login request")
}

// AcceptLoginRequest accepts the requested authentication process, and returns redirect URI.
func (lrd *LoginReqDoer) AcceptLoginRequest(challenge string, remember bool, subject string) (string, error) {
	data := struct {
		Remember    bool     `json:"remember"`
		RememberFor int      `json:"remember_for"`
		Subject     string   `json:"subject"`
		ACR         string   `json:"acr,omitempty"`
		AMR         []string `json:"amr,omitempty"`
	}{
		Remember:    remember,
		RememberFor: lrd.rememberFor,
		Subject:     subject,
		ACR:         lrd.acr,
		AMR:         lrd.amr,
	}
	redirectURI, err := acceptRequest(login, lrd.hydraURL, lrd.fakeTLSTermination, challenge, data)
	return redirectURI, errors.Wrap(err, "failed to accept login request")
}
