//go:generate go-bindata -pkg lgraphql -o lgraphql/lgraphql.go -nometadata _lgraphql/

// Package client implements the interfaces required by the parent lagoon
// package.
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amazeeio/lagoon-cli/internal/lagoon/client/lgraphql"
	"github.com/machinebox/graphql"
)

const version = "0.x.x"

// Client implements the lagoon package interfaces for the Lagoon GraphQL API.
type Client struct {
	userAgent string
	token     string
	client    *graphql.Client
}

// New creates a new Client for the given endpoint.
func New(endpoint, token string, debug bool) *Client {
	if debug {
		return &Client{
			token: token,
			client: graphql.NewClient(endpoint,
				// enable debug logging to stderr
				func(c *graphql.Client) {
					l := log.New(os.Stderr, "graphql", 0)
					c.Log = func(s string) {
						l.Println(s)
					}
				}),
		}
	}
	return &Client{
		token:  token,
		client: graphql.NewClient(endpoint),
	}
}

// newRequest constructs a graphql request.
// assetName is the name of the graphql query template in _graphql/.
// varStruct is converted to a map of variables for the template.
func (c *Client) newRequest(
	assetName string, varStruct interface{}) (*graphql.Request, error) {

	q, err := lgraphql.Asset(assetName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get asset: %w", err)
	}

	vars, err := structToVarMap(varStruct)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert struct to map: %w", err)
	}

	req := graphql.NewRequest(string(q))
	for key, value := range vars {
		req.Var(key, value)
	}

	headers := map[string]string{
		"User-Agent":    fmt.Sprintf("lagoon-cli version: %s", version),
		"Authorization": fmt.Sprintf("Bearer %s", c.token),
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// structToVarMap encodes the given struct to a map. The idea is that by
// round-tripping through Marshal/Unmarshal, omitempty is applied to the
// zero-valued fields.
func structToVarMap(
	varStruct interface{}) (vars map[string]interface{}, err error) {
	data, err := json.Marshal(varStruct)
	if err != nil {
		return vars, err
	}
	return vars, json.Unmarshal(data, &vars)
}
