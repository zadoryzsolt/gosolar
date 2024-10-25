package gosolar

import (
	"bytes"
	"crypto/tls"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Client structure for the SolarWinds (SWIS) connection.
type Client struct {
	// connection parameters
	URL      string
	Username string
	Password string

	// internal state
	http *http.Client
}

// NewClient creates a new reference to the Client struct.
func NewClient(host, user, pass string, ignoreSSL bool) *Client {
	return &Client{
		URL:      fmt.Sprintf("https://%s:17778/SolarWinds/InformationService/v3/Json/", host),
		Username: user,
		Password: pass,
		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: ignoreSSL,
				},
				MaxIdleConnsPerHost: 4, // DEFAULT: 2
			},
		},
	}
}

func (c *Client) post(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL+endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.Username, c.Password)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit query: %v", err)
	}

	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("query failed - status code %d: %v", res.StatusCode, err)
	}
	res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("swis failure message [status: %d]:\n%s",
			res.StatusCode, string(output))
	}

	return output, nil
}

func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.URL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.Username, c.Password)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit query: %v", err)
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		err2 := res.Body.Close()
		if err2 != nil {
			log.Fatalf("failed to close result body: %v", err2)
		}
	}()

	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("query failed - status code %d: %v", res.StatusCode, err)
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("swis failure message [status: %d]:\n%s",
			res.StatusCode, string(output))
	}

	return output, nil
}

// Query retrieves a result from the SolarWinds API.
func (c *Client) Query(ctx context.Context, query string, parameters interface{}) ([]byte, error) {
	req := struct {
		Query      string      `json:"query"`
		Parameters interface{} `json:"parameters"`
	}{
		Query:      query,
		Parameters: parameters,
	}

	result, err := c.post(ctx, "Query", &req)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}

	sr := struct {
		Result *json.RawMessage `json:"results"`
	}{}

	err = json.Unmarshal(result, &sr)
	if err != nil {
		return nil, err
	}

	return []byte(*sr.Result), nil
}

// QueryOne wraps QueryRow which wraps post and extracts a single value.
func (c *Client) QueryOne(ctx context.Context, query string, parameters interface{}) (interface{}, error) {
	res, err := c.QueryRow(ctx, query, parameters)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})

	if json.Unmarshal(res, &m); err != nil {
		return nil, fmt.Errorf("could not unmarshal the result: %v", err)
	}

	var value interface{}
	for _, v := range m {
		value = v
		break
	}

	return value, nil
}

// QueryRow wraps query and pulls a single row from the result.
func (c *Client) QueryRow(ctx context.Context, query string, parameters interface{}) ([]byte, error) {
	res, err := c.Query(ctx, query, parameters)
	if err != nil {
		return nil, err
	}

	return res[1 : len(res)-1], nil
}

// QueryColumn wraps Query and pulls a single column of values into a slice of maps.
func (c *Client) QueryColumn(ctx context.Context, query string, parameters interface{}) ([]interface{}, error) {
	res, err := c.Query(ctx, query, parameters)
	if err != nil {
		return nil, err
	}

	var rows []map[string]interface{}

	if json.Unmarshal(res, &rows); err != nil {
		return nil, fmt.Errorf("could not unmarshal the result: %v", err)
	}

	var values []interface{}
	for _, m := range rows {
		for _, v := range m {
			values = append(values, v)
		}
	}

	return values, nil
}

// Create calls the create endpoint and passes the entity and body.
func (c *Client) Create(ctx context.Context, entity, body interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("Create/%s", entity)

	return c.post(ctx, endpoint, body)
}

func (c *Client) Read(ctx context.Context, uri string) ([]byte, error) {
	return c.get(ctx, uri)
}

// Invoke calls the invoke endpoint with the entity and verb along with a body.
func (c *Client) Invoke(ctx context.Context, entity, verb string, body interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("Invoke/%s/%s", entity, verb)

	return c.post(ctx, endpoint, body)
}

// BulkDelete wraps post and send a slice of URIs to delete.
func (c *Client) BulkDelete(ctx context.Context, uris []string) ([]byte, error) {
	req := map[string][]string{
		"uris": uris,
	}

	return c.post(ctx, "BulkDelete", req)
}

// Delete wraps post and uses the DELETE method to delete an entity.
func (c *Client) Delete(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.URL+uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new request: %v", err)
	}

	req.SetBasicAuth(c.Username, c.Password)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete: %v", err)
	}

	output, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("delete failed - status code %d: %v", res.StatusCode, err)
	}
	res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("swis failure message [status: %d]:\n%s",
			res.StatusCode, string(output))
	}

	return output, nil
}

// Update wraps the post function passing the URI and body to update.
func (c *Client) Update(ctx context.Context, uri string, body map[string]interface{}) ([]byte, error) {
	return c.post(ctx, uri, body)
}
