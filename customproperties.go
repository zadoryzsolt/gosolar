package gosolar

import (
	"context"
	"fmt"
)

// BulkSetCustomProperty sets a custom property on a series of URIs.
func (c *Client) BulkSetCustomProperty(ctx context.Context, uris []string, name string, value interface{}) error {
	// load up the uris that are going to be affected
	var cpuris []string
	for _, uri := range uris {
		cpuris = append(cpuris, uri+"/CustomProperties")
	}

	bulkRequest := struct {
		URIs       []string               `json:"uris"`
		Properties map[string]interface{} `json:"properties"`
	}{
		URIs: cpuris,
		Properties: map[string]interface{}{
			name: value,
		},
	}

	_, err := c.post(ctx, "BulkUpdate", &bulkRequest)
	if err != nil {
		return fmt.Errorf("failed to post bulk update: %v", err)
	}

	return nil
}

// SetCustomProperty sets a custom property value on a specific URI.
func (c *Client) SetCustomProperty(ctx context.Context, uri, name string, value interface{}) error {
	property := map[string]interface{}{
		name: value,
	}

	_, err := c.post(ctx, uri+"/CustomProperties", &property)
	if err != nil {
		return fmt.Errorf("failed to update custom property: %v", err)
	}

	return nil
}

// SetCustomProperties sets multiple properties on an entity.
func (c *Client) SetCustomProperties(ctx context.Context, uri string, properties map[string]interface{}) error {
	_, err := c.post(ctx, uri+"/CustomProperties", &properties)
	if err != nil {
		return fmt.Errorf("failed to update custom property: %v", err)
	}

	return nil
}

// CreateCustomProperty creates a new custom property of a specified type.
func (c *Client) CreateCustomProperty(ctx context.Context, cpEntity, cpType, cpName, cpDesc string) error {
	var cpLength string

	if cpType == "string" {
		cpLength = "400"
	} else {
		cpLength = "0"
	}

	props := []string{
		cpName,
		cpDesc,
		cpType,
		cpLength,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"false",
		"",
	}

	endpoint := fmt.Sprintf("Invoke/%s/CreateCustomProperty", cpEntity)

	_, err := c.post(ctx, endpoint, &props)
	if err != nil {
		return fmt.Errorf("failed to create custom property: %v", err)
	}

	return nil
}
