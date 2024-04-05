package httpclient

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
)

type ClientConfig struct {
	RootURL              string
	APIKey               string
	mTLSCertificatePaths *apihelpers.CertificatePaths
	Timeout              time.Duration
}

func (cConfig ClientConfig) RunHTTPcall(pathname string, payload interface{}) (map[string]interface{}, error) {
	json_data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	transport, err := getTransportWithMTLSConfig(cConfig.mTLSCertificatePaths)
	if err != nil {
		slog.Error("Error creating transport with mTLS config", slog.String("error", err.Error()))
		return nil, err
	}

	client := &http.Client{
		Timeout: cConfig.Timeout,
	}
	if transport != nil {
		client.Transport = transport
	}

	url := cConfig.RootURL + pathname
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("unexpected error in preparing http request", slog.String("error", err.Error()))
		return nil, err
	}
	if cConfig.APIKey != "" {
		req.Header.Set("Api-Key", cConfig.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("unexpected error in http call", slog.String("error", err.Error()))
		return nil, err
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		slog.Error("Error decoding response", slog.String("error", err.Error()))
		return nil, err
	}
	return res, nil
}

func getTransportWithMTLSConfig(mTLSCertificatePaths *apihelpers.CertificatePaths) (*http.Transport, error) {
	if mTLSCertificatePaths == nil {
		return nil, nil
	}

	tlsConfig, err := apihelpers.LoadTLSConfig(*mTLSCertificatePaths)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}, nil
}
