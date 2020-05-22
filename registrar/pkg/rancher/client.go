package rancher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type ClusterRegistrationTokenResponse struct {
	Type         string                         `json:"type"`
	Links        Links                          `json:"links"`
	CreateTypes  map[string]string              `json:"createTypes"`
	Actions      map[string]string              `json:"actions"`
	Pagination   Pagination                     `json:"pagination"`
	Sort         Sort                           `json:"sort"`
	Filters      Filters                        `json:"filters"`
	ResourceType string                         `json:"resourceType"`
	Data         []ClusterRegistrationTokenData `json:"data"`
}
type ClusterRegistrationTokenLinks struct {
	Self string `json:"self"`
}
type Pagination struct {
	Limit int `json:"limit"`
	Total int `json:"total"`
}
type ClusterRegistrationTokenDataLinks struct {
	Command              string `json:"command"`
	InsecureCommand      string `json:"insecureCommand"`
	ManifestURL          string `json:"manifestUrl"`
	NodeCommand          string `json:"nodeCommand"`
	State                string `json:"state"`
	Token                string `json:"token"`
	Transitioning        string `json:"transitioning"`
	TransitioningMessage string `json:"transitioningMessage"`
	UUID                 string `json:"uuid"`
	WindowsNodeCommand   string `json:"windowsNodeCommand"`
}
type Sort struct {
	Order   string `json:"order"`
	Reverse string `json:"reverse"`
	Links   Links  `json:"links"`
}
type Filters struct {
	ClusterID            interface{} `json:"clusterId"`
	Command              interface{} `json:"command"`
	Created              interface{} `json:"created"`
	CreatorID            interface{} `json:"creatorId"`
	InsecureCommand      interface{} `json:"insecureCommand"`
	ManifestURL          interface{} `json:"manifestUrl"`
	Name                 interface{} `json:"name"`
	NamespaceID          interface{} `json:"namespaceId"`
	NodeCommand          interface{} `json:"nodeCommand"`
	Removed              interface{} `json:"removed"`
	State                interface{} `json:"state"`
	Token                interface{} `json:"token"`
	Transitioning        interface{} `json:"transitioning"`
	TransitioningMessage interface{} `json:"transitioningMessage"`
	UUID                 interface{} `json:"uuid"`
	WindowsNodeCommand   interface{} `json:"windowsNodeCommand"`
}
type Links struct {
	Remove string `json:"remove"`
	Self   string `json:"self"`
	Update string `json:"update"`
}
type ClusterRegistrationTokenData struct {
	Annotations          map[string]string                 `json:"annotations"`
	BaseType             string                            `json:"baseType"`
	ClusterID            string                            `json:"clusterId"`
	Command              string                            `json:"command"`
	Created              time.Time                         `json:"created"`
	CreatedTS            int64                             `json:"createdTS"`
	CreatorID            string                            `json:"creatorId"`
	ID                   string                            `json:"id"`
	InsecureCommand      string                            `json:"insecureCommand"`
	Labels               map[string]string                 `json:"labels"`
	Links                ClusterRegistrationTokenDataLinks `json:"links"`
	ManifestURL          string                            `json:"manifestUrl"`
	Name                 string                            `json:"name"`
	NamespaceID          interface{}                       `json:"namespaceId"`
	NodeCommand          string                            `json:"nodeCommand"`
	State                string                            `json:"state"`
	Token                string                            `json:"token"`
	Transitioning        string                            `json:"transitioning"`
	TransitioningMessage string                            `json:"transitioningMessage"`
	Type                 string                            `json:"type"`
	UUID                 string                            `json:"uuid"`
	WindowsNodeCommand   string                            `json:"windowsNodeCommand"`
}

// Client is a rancher client
type Client struct {
	h       *http.Client
	authKey string
	baseURL *url.URL
}

func NewClient(authKey string) *Client {
	u, err := url.Parse("https://rancher.tritonjs.com")
	if err != nil {
		panic(err)
	}

	return &Client{
		h:       &http.Client{},
		authKey: authKey,
		// TODO(jaredallard): fix this hardcode
		baseURL: u,
	}
}

// GetClusterRegistrationToken returns all cluster registration tokens or, if clusterId is provided
// all tokens for a given server.
func (c *Client) GetClusterRegistrationToken(ctx context.Context, clusterId string) ([]ClusterRegistrationTokenData, error) {
	u := url.URL{
		Scheme: c.baseURL.Scheme,
		Host:   c.baseURL.Host,
		Path:   "/v3/clusterregistrationtokens",
	}

	if clusterId != "" {
		q := u.Query()
		q.Set("clusterId", clusterId)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.authKey))

	resp, err := c.h.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	// handle any errors that pop up
	if resp.StatusCode != http.StatusOK {
		raw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			raw = []byte("failed to read body")
		}
		return nil, fmt.Errorf("got non 200 status code %d: %s", resp.StatusCode, string(raw))
	}

	var crt ClusterRegistrationTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&crt); err != nil {
		return nil, errors.Wrap(err, "failed to parse body")
	}

	return crt.Data, nil
}
