package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/publicip/types"
)

type ipLocation struct {
	client *http.Client
	token  string
}

func newIPLocation(client *http.Client, token string) *ipLocation {
	return &ipLocation{
		client: client,
		token:  token,
	}
}

// FetchInfo obtains information on the ip address provided
// using the api.iplocation.net API. If the ip is the zero value,
// the public IP address of the machine is used as the IP.
func (i *ipLocation) FetchInfo(ctx context.Context, logger types.Logger, ip netip.Addr) (
	result models.PublicIP, err error) {
	url := "https://api.iplocation.net/"
	if !ip.IsValid() {
		url += "?cmd=get-ip"

		logger.Info("ip address is not valid, requesting external ip address")
		logger.Info(url)
		
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			logger.Error("request failed")
			return result, err
		}

		response, err := i.client.Do(request)
		if err != nil {
			logger.Error("request execute failed")
			return result, err
		}

		defer response.Body.Close()

	        switch response.StatusCode {
        	case http.StatusOK:
	        case http.StatusTooManyRequests, http.StatusForbidden:
        	        return result, fmt.Errorf("%w from %s: %d %s",
                	        ErrTooManyRequests, url, response.StatusCode, response.Status)
	        default:
        	        return result, fmt.Errorf("%w from %s: %d %s",
                	        ErrBadHTTPStatus, url, response.StatusCode, response.Status)
		}

	        decoder := json.NewDecoder(response.Body)
        	var data struct {
                	IP          netip.Addr `json:"ip,omitempty"`
        	}
	        if err := decoder.Decode(&data); err != nil {
      	        	return result, fmt.Errorf("decoding get-ip response: %w", err)
        	}			

		ip = data.IP
		url = "https://api.iplocation.net/"
        }

	if ip.IsValid() {
		url += "?ip=" + ip.String()
	}

	if i.token != "" {
		if !strings.Contains(url, "?") {
			url += "?"
		}
		url += "&key=" + i.token
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return result, err
	}

	response, err := i.client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	if i.token != "" && response.StatusCode == http.StatusUnauthorized {
		return result, fmt.Errorf("%w: %s", ErrTokenNotValid, response.Status)
	}

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests, http.StatusForbidden:
		return result, fmt.Errorf("%w from %s: %d %s",
			ErrTooManyRequests, url, response.StatusCode, response.Status)
	default:
		return result, fmt.Errorf("%w from %s: %d %s",
			ErrBadHTTPStatus, url, response.StatusCode, response.Status)
	}

	decoder := json.NewDecoder(response.Body)
	var data struct {
		IP          netip.Addr `json:"ip,omitempty"`
		CountryName string     `json:"country_name,omitempty"`
		Org         string     `json:"isp,omitempty"`
	}
	if err := decoder.Decode(&data); err != nil {
		return result, fmt.Errorf("decoding response: %w", err)
	}

	result = models.PublicIP{
		IP:           data.IP,
		Country:      data.CountryName,
		Organization: data.Org,
	}
	return result, nil
}
