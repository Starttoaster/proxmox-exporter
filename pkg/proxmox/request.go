package proxmox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

// NewRequest creates a new request.
// Method should be a valid http request method.
// Path should be an API path relative to the client's base URL.
// Path should not have a preceding '/'
// If specified, the value pointed to by opt is encoded into the query string of the URL.
func (c *Client) NewRequest(method, path string, opt interface{}) (*http.Request, error) {
	u := *c.baseURL
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	// Set the encoded path data
	u.RawPath = c.baseURL.Path + path
	u.Path = c.baseURL.Path + unescaped

	// Set query parameters if any are provided
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	// Create request
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set request header if making a POST or PUT
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		req.Header.Set("Content-Type", "x-www-form-urlencoded")
	}

	return req, nil
}

// Do sends an API request. The response is stored in the value 'v' or returned as an error.
// If v implements the io.Writer interface, the raw response body will be written to v, without json decoding it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// Set auth header according to https://pve.proxmox.com/wiki/Proxmox_VE_API#Authentication
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.token))

	// Do request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	// Check for error API response and capture it as an error
	if resp.StatusCode > 399 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading Proxmox response body: %v", err)
		}

		return resp, fmt.Errorf(string(body))
	}

	// Copy body into v
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return resp, err
}
