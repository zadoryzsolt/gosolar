package gosolar

import (
    "fmt"
    "context"
)

// RemoveNCMNodes deletes nodes from NCM handling in SolarWinds.
func (c *Client) RemoveNCMNodes(ctx context.Context, guids []string) error {
	endpoint := "Invoke/Cirrus.Nodes/RemoveNodes"
	req := [][]string{guids}

	_, err := c.post(ctx, endpoint, req)

	if err != nil {
		return fmt.Errorf("failed to remove the NCM nodes %v", err)
	}

	return nil
}
