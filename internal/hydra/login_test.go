package hydra_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/i-core/werther/internal/hydra"
	"github.com/pkg/errors"
)

func TestInitiateLoginRequest(t *testing.T) {
	testCases := []struct {
		name      string
		challenge string
		reqInfo   *hydra.ReqInfo
		status    int
		wantErr   error
	}{
		{
			name:    "challenge is missed",
			wantErr: hydra.ErrChallengeMissed,
		},
		{
			name:      "unauthenticated",
			challenge: "foo",
			status:    401,
			wantErr:   hydra.ErrUnauthenticated,
		},
		{
			name:      "challenge is not found",
			challenge: "foo",
			status:    404,
			wantErr:   hydra.ErrChallengeNotFound,
		},
		{
			name:      "challenge is expired",
			challenge: "foo",
			status:    409,
			wantErr:   hydra.ErrChallengeExpired,
		},
		{
			name:      "happy path",
			challenge: "foo",
			status:    200,
			reqInfo: &hydra.ReqInfo{
				Challenge:       "foo",
				RequestedScopes: []string{"profile", "email"},
				Skip:            true,
				Subject:         "testSubject",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &testInitiateLoginHandler{reqInfo: tc.reqInfo, status: tc.status}
			srv := httptest.NewServer(h)
			defer srv.Close()
			ldr := hydra.NewLoginReqDoer(srv.URL, false, 0, "", nil)

			reqInfo, err := ldr.InitiateRequest(tc.challenge)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("\ngot no errors\nwant error:\n\t%s", tc.wantErr)
				}
				err = errors.Cause(err)
				if err != tc.wantErr {
					t.Fatalf("\ngot error:\n\t%s\nwant error:\n\t%s", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("\ngot error:\n\t%s\nwant no errors", err)
			}

			if h.challenge != tc.challenge {
				t.Errorf("\ngot challenge:\n\t%#v\nwant challenge:\n\t%#v", h.challenge, tc.challenge)
			}
			if !reflect.DeepEqual(tc.reqInfo, reqInfo) {
				t.Errorf("\ngot request info:\n\t%#v\nwant request info:\n\t%#v", reqInfo, tc.reqInfo)
			}
		})
	}
}

type testInitiateLoginHandler struct {
	reqInfo   *hydra.ReqInfo
	status    int
	challenge string
}

func (h *testInitiateLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/oauth2/auth/requests/login" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"error": http.StatusText(http.StatusMethodNotAllowed)}); err != nil {
			panic(fmt.Sprintf("initial request: failed to write response: %s", err))
		}
		return
	}
	h.challenge = r.URL.Query().Get("login_challenge")
	w.WriteHeader(h.status)
	if h.status == http.StatusOK {
		if err := json.NewEncoder(w).Encode(h.reqInfo); err != nil {
			panic(fmt.Sprintf("initial request: failed to write response: %s", err))
		}
	}
}

func TestAcceptLoginRequest(t *testing.T) {
	testCases := []struct {
		name        string
		challenge   string
		rememberFor int
		remember    bool
		subject     string
		status      int
		redirect    string
		wantErr     error
	}{
		{
			name:    "challenge is missed",
			wantErr: hydra.ErrChallengeMissed,
		},
		{
			name:        "unauthenticated",
			challenge:   "foo",
			rememberFor: 10,
			remember:    true,
			subject:     "testSubject",
			status:      http.StatusUnauthorized,
			wantErr:     hydra.ErrUnauthenticated,
		},
		{
			name:        "challenge is not found",
			challenge:   "foo",
			rememberFor: 10,
			remember:    true,
			subject:     "testSubject",
			status:      http.StatusNotFound,
			wantErr:     hydra.ErrChallengeNotFound,
		},
		{
			name:        "happy path",
			challenge:   "foo",
			rememberFor: 10,
			remember:    true,
			subject:     "testSubject",
			status:      http.StatusOK,
			redirect:    "/test-redirect",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &testAcceptLoginHandler{challenge: tc.challenge, status: tc.status, redirect: tc.redirect}
			srv := httptest.NewServer(h)
			defer srv.Close()
			ldr := hydra.NewLoginReqDoer(srv.URL, false, tc.rememberFor, "", nil)

			redirect, err := ldr.AcceptLoginRequest(tc.challenge, tc.remember, tc.subject)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("\ngot no errors\nwant error:\n\t%s", tc.wantErr)
				}
				err = errors.Cause(err)
				if err.Error() != tc.wantErr.Error() {
					t.Fatalf("\ngot error:\n\t%s\nwant error:\n\t%s", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("\ngot error:\n\t%s\nwant no errors", err)
			}

			if h.challenge != tc.challenge {
				t.Errorf("\ngot challenge:\n\t%#v\nwant challenge:\n\t%#v", h.challenge, tc.challenge)
			}
			wantData := map[string]interface{}{
				"remember":     tc.remember,
				"remember_for": tc.rememberFor,
				"subject":      tc.subject,
			}
			if !reflect.DeepEqual(h.data, wantData) {
				t.Errorf("\ngot request data:\n\t%#v\nwant request data:\n\t%#v", h.data, wantData)
			}
			if redirect != tc.redirect {
				t.Errorf("\ngot redirect URL:\n\t%#v\nwant redirect URL:\n\t%#v", redirect, tc.redirect)
			}
		})
	}
}

type testAcceptLoginHandler struct {
	challenge string
	data      map[string]interface{}
	status    int
	redirect  string
}

func (h *testAcceptLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut || r.URL.Path != "/oauth2/auth/requests/login/accept" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	h.challenge = r.URL.Query().Get("login_challenge")
	w.WriteHeader(h.status)
	if r.Body != http.NoBody {
		// Note: Go JSON Decoder decodes numbers as float64, but we need int.
		// So we convert numbers to int manually.
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			panic(fmt.Sprintf("accept request: failed to read request body: %s", err))
		}
		h.data = make(map[string]interface{}, len(raw))
		for key, val := range raw {
			s := string(val)
			if i, err := strconv.Atoi(s); err == nil {
				h.data[key] = i
				continue
			}
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				h.data[key] = f
				continue
			}
			var v interface{}
			if err := json.Unmarshal(val, &v); err == nil {
				h.data[key] = v
				continue
			}
			h.data[key] = val
		}
	}
	if h.status == http.StatusOK {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"redirect_to": h.redirect}); err != nil {
			panic(fmt.Sprintf("accept request: failed to write response: %s", err))
		}
	}
}
