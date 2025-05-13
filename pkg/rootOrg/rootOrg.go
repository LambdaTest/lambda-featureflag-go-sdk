package rootOrg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LambdaTest/lambda-featureflag-go-sdk/model"
)

// lumsClient handles LUMS API operations
type lumsClient struct {
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
}

// NewLumsClient creates a new LUMSClient
func newLumsClient(httpClient *http.Client, url string, timeout time.Duration) *lumsClient {
	return &lumsClient{
		httpClient: httpClient,
		baseURL:    url,
		timeout:    timeout,
	}
}

func (c *lumsClient) getRootOrgs(apiKey string) (*model.OrgMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	url := fmt.Sprintf("%s/sdk/rootOrg", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", apiKey))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %v", err)
		}
		return nil, fmt.Errorf("%s: %s", errResp.Title, errResp.Message)
	}

	var successResp model.SuccessResponse
	err = json.NewDecoder(resp.Body).Decode(&successResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode success response: %v", err)
	}

	return &successResp.Data, nil
}

// GetRootOrgs is a convenience function that uses the default HTTP client
func GetRootOrgs(apiKey string, url string, timeout time.Duration) (*model.OrgMap, error) {
	client := newLumsClient(http.DefaultClient, url, timeout)
	result, err := client.getRootOrgs(apiKey)
	if err != nil {
		return nil, err
	}
	return result, nil
}
