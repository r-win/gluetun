package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
)

type ipApi struct {
	client *http.Client
	token  string
}

func newIPApi(client *http.Client, token string) *ipApi {
	return &ipApi{
		client: client,
		token:  token,
	}
}

// FetchInfo obtains information on the ip address provided
// using the ipapi.co API. If the ip is the zero value, the public IP address
// of the machine is used as the IP.
func (i *ipApi) FetchInfo(ctx context.Context, ip netip.Addr) (
	result models.PublicIP, err error) {
	url := "https://ipapi.co"
	if ip.IsValid() {
		url += "/ip=" + ip.String()
	}
	url += "/json"

	if i.token != "" {
		url += "?key=" + i.token
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
		IP       netip.Addr `json:"ip,omitempty"`
		Region   string     `json:"region,omitempty"`
		Country  string     `json:"country_name,omitempty"`
		City     string     `json:"city,omitempty"`
		Hostname string     `json:"hostname,omitempty"`
		Lat      string     `json:"latitude,omitempty"`
		Lng      string     `json:"longitude,omitempty"`
		Org      string     `json:"org,omitempty"`
		Postal   string     `json:"postal,omitempty"`
		Timezone string     `json:"timezone,omitempty"`
	}
	if err := decoder.Decode(&data); err != nil {
		return result, fmt.Errorf("decoding response: %w", err)
	}

	countryCode := strings.ToLower(data.Country)
	country, ok := constants.CountryCodes()[countryCode]
	if ok {
		data.Country = country
	}

	result = models.PublicIP{
		IP:           data.IP,
		Region:       data.Region,
		Country:      data.Country,
		City:         data.City,
		Hostname:     data.Hostname,
		Location:     data.Lat + "," + data.Lng,
		Organization: data.Org,
		PostalCode:   data.Postal,
		Timezone:     data.Timezone,
	}
	return result, nil
}
