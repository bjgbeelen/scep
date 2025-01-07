package scepserver

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"errors"
	"strings"

	"github.com/smallstep/scep"
)

type DeviceChallenge struct {
	DeviceId string  `json:"device_id"`
	Challenge string `json:"challenge"`
}

// CSRSignerContext is a handler for signing CSRs by a CA/RA.
//
// SignCSRContext should take the CSR in the CSRReqMessage and return a
// Certificate signed by the CA.
type CSRSignerContext interface {
	SignCSRContext(context.Context, *scep.CSRReqMessage) (*x509.Certificate, error)
}

// CSRSignerContextFunc is an adapter for CSR signing by the CA/RA.
type CSRSignerContextFunc func(context.Context, *scep.CSRReqMessage) (*x509.Certificate, error)

// SignCSR calls f(ctx, m).
func (f CSRSignerContextFunc) SignCSRContext(ctx context.Context, m *scep.CSRReqMessage) (*x509.Certificate, error) {
	return f(ctx, m)
}

// CSRSigner is a handler for CSR signing by the CA/RA
//
// SignCSR should take the CSR in the CSRReqMessage and return a
// Certificate signed by the CA.
type CSRSigner interface {
	SignCSR(*scep.CSRReqMessage) (*x509.Certificate, error)
}

// CSRSignerFunc is an adapter for CSR signing by the CA/RA.
type CSRSignerFunc func(*scep.CSRReqMessage) (*x509.Certificate, error)

// SignCSR calls f(m).
func (f CSRSignerFunc) SignCSR(m *scep.CSRReqMessage) (*x509.Certificate, error) {
	return f(m)
}

// NopCSRSigner does nothing.
func NopCSRSigner() CSRSignerContextFunc {
	return func(_ context.Context, _ *scep.CSRReqMessage) (*x509.Certificate, error) {
		return nil, nil
	}
}

func ExternalChallengeMiddleware(url string, next CSRSignerContext ) CSRSignerContextFunc {
	return func(ctx context.Context, m *scep.CSRReqMessage) (*x509.Certificate, error) {
		deviceId := strings.Join(m.CSR.Subject.OrganizationalUnit, ", ")

		deviceChallenge := DeviceChallenge{
			DeviceId: deviceId,
			Challenge: m.ChallengePassword,
		}
		jsonData, err := json.Marshal(deviceChallenge)
		if err != nil {
			return nil, errors.New("Unable to marshall device challenge")
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, errors.New("Failed to test external challenge")
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return next.SignCSRContext(ctx, m)
		} else {
			return nil, errors.New("Invalid challenge")
		}
	}
}

// StaticChallengeMiddleware wraps next and validates the challenge from the CSR.
func StaticChallengeMiddleware(challenge string, next CSRSignerContext) CSRSignerContextFunc {
	challengeBytes := []byte(challenge)
	return func(ctx context.Context, m *scep.CSRReqMessage) (*x509.Certificate, error) {
		// TODO: compare challenge only for PKCSReq?
		if subtle.ConstantTimeCompare(challengeBytes, []byte(m.ChallengePassword)) != 1 {
			return nil, errors.New("invalid challenge")
		}

		return next.SignCSRContext(ctx, m)
	}
}

// SignCSRAdapter adapts a next (i.e. no context) to a context signer.
func SignCSRAdapter(next CSRSigner) CSRSignerContextFunc {
	return func(_ context.Context, m *scep.CSRReqMessage) (*x509.Certificate, error) {
		return next.SignCSR(m)
	}
}
