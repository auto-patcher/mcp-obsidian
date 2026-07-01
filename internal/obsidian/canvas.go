package obsidian

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// --- Canvas file operations -------------------------------------------------

// ListCanvasFiles lists canvas files in the vault root or a subdirectory.
func (c *Client) ListCanvasFiles(dirpath string) (json.RawMessage, error) {
	path := "/canvas/"
	if dirpath != "" {
		path = "/canvas/" + dirpath + "/"
	}
	data, err := c.do(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetCanvas returns the full canvas JSON (nodes and edges).
func (c *Client) GetCanvas(filepath string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// WriteCanvas creates or overwrites a canvas file.
func (c *Client) WriteCanvas(filepath string, canvasData any) (json.RawMessage, error) {
	body, err := json.Marshal(canvasData)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(http.MethodPut, "/canvas/"+filepath,
		map[string]string{"Content-Type": "application/json"},
		body)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return json.RawMessage(`{"status":"ok"}`), nil
	}
	return json.RawMessage(resp), nil
}

// DeleteCanvas deletes a canvas file.
func (c *Client) DeleteCanvas(filepath string) error {
	_, err := c.do(http.MethodDelete, "/canvas/"+filepath, nil, nil)
	return err
}

// SearchCanvases searches canvas files by query, optionally scoped to a directory.
func (c *Client) SearchCanvases(query, dirpath string) (json.RawMessage, error) {
	path := "/canvas/search/?query=" + url.QueryEscape(query)
	if dirpath != "" {
		path += "&dirPath=" + url.QueryEscape(dirpath)
	}
	data, err := c.do(http.MethodPost, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetCanvasStats returns node/edge counts and bounding box for a canvas.
func (c *Client) GetCanvasStats(filepath string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/stats", nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// --- Node operations --------------------------------------------------------

// ListCanvasNodes lists nodes in a canvas, optionally filtered by type.
func (c *Client) ListCanvasNodes(filepath, nodeType string) (json.RawMessage, error) {
	path := "/canvas/" + filepath + "/nodes"
	if nodeType != "" {
		path += "?type=" + url.QueryEscape(nodeType)
	}
	data, err := c.do(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetCanvasNode returns a specific node by ID.
func (c *Client) GetCanvasNode(filepath, nodeID string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/nodes/"+nodeID, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// CreateCanvasNode creates a new node and returns it with its generated ID.
func (c *Client) CreateCanvasNode(filepath string, node any) (json.RawMessage, error) {
	return c.jsonPOST("/canvas/"+filepath+"/nodes", node)
}

// UpdateCanvasNode updates fields on an existing node.
func (c *Client) UpdateCanvasNode(filepath, nodeID string, updates any) (json.RawMessage, error) {
	return c.jsonPUT("/canvas/"+filepath+"/nodes/"+nodeID, updates)
}

// DeleteCanvasNode deletes a node; optionally removes its connected edges too.
func (c *Client) DeleteCanvasNode(filepath, nodeID string, deleteEdges bool) error {
	path := "/canvas/" + filepath + "/nodes/" + nodeID
	if deleteEdges {
		path += "?deleteEdges=true"
	}
	_, err := c.do(http.MethodDelete, path, nil, nil)
	return err
}

// --- Edge operations --------------------------------------------------------

// ListCanvasEdges lists all edges in a canvas.
func (c *Client) ListCanvasEdges(filepath string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/edges", nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetCanvasEdge returns a specific edge by ID.
func (c *Client) GetCanvasEdge(filepath, edgeID string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/edges/"+edgeID, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// CreateCanvasEdge creates a new edge and returns it with its generated ID.
func (c *Client) CreateCanvasEdge(filepath string, edge any) (json.RawMessage, error) {
	return c.jsonPOST("/canvas/"+filepath+"/edges", edge)
}

// UpdateCanvasEdge updates fields on an existing edge.
func (c *Client) UpdateCanvasEdge(filepath, edgeID string, updates any) (json.RawMessage, error) {
	return c.jsonPUT("/canvas/"+filepath+"/edges/"+edgeID, updates)
}

// DeleteCanvasEdge deletes an edge by ID.
func (c *Client) DeleteCanvasEdge(filepath, edgeID string) error {
	_, err := c.do(http.MethodDelete, "/canvas/"+filepath+"/edges/"+edgeID, nil, nil)
	return err
}

// --- Group operations -------------------------------------------------------

// ListCanvasGroups lists all group nodes in a canvas.
func (c *Client) ListCanvasGroups(filepath string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/groups", nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// GetCanvasGroup returns a group node and the nodes it contains spatially.
func (c *Client) GetCanvasGroup(filepath, groupID string) (json.RawMessage, error) {
	data, err := c.do(http.MethodGet, "/canvas/"+filepath+"/groups/"+groupID, nil, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// CreateCanvasGroup creates a new group node.
func (c *Client) CreateCanvasGroup(filepath string, group any) (json.RawMessage, error) {
	return c.jsonPOST("/canvas/"+filepath+"/groups", group)
}

// UpdateCanvasGroup updates fields on an existing group node.
func (c *Client) UpdateCanvasGroup(filepath, groupID string, updates any) (json.RawMessage, error) {
	return c.jsonPUT("/canvas/"+filepath+"/groups/"+groupID, updates)
}

// DeleteCanvasGroup deletes a group node by ID.
func (c *Client) DeleteCanvasGroup(filepath, groupID string) error {
	_, err := c.do(http.MethodDelete, "/canvas/"+filepath+"/groups/"+groupID, nil, nil)
	return err
}

// --- shared helpers ---------------------------------------------------------

func (c *Client) jsonPOST(path string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	data, err := c.do(http.MethodPost, path,
		map[string]string{"Content-Type": "application/json"},
		body)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (c *Client) jsonPUT(path string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	data, err := c.do(http.MethodPut, path,
		map[string]string{"Content-Type": "application/json"},
		body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return json.RawMessage(`{"status":"ok"}`), nil
	}
	return json.RawMessage(data), nil
}
