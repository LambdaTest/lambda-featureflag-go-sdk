package rootOrg

import (
	"strconv"
	"sync"

	"github.com/LambdaTest/lambda-featureflag-go-sdk/model"
)

type Client struct {
	apiKey     string
	Config     *Config
	poller     *poller
	flags      *model.OrgMap
	flagsMutex *sync.RWMutex
}

func NewClient(apiKey string, config *Config) *Client {
	return &Client{
		apiKey:     apiKey,
		Config:     config,
		flagsMutex: &sync.RWMutex{},
		flags:      &model.OrgMap{},
		poller:     newPoller(),
	}
}

func (c *Client) Start() error {
	c.poller = newPoller()
	c.poller.Poll(c.Config.FlagConfigPollerInterval, func() {
		c.pollFlags()
	})
	return c.pollFlags()
}

func (c *Client) Stop() {
	close(c.poller.shutdown)
}

func (c *Client) pollFlags() error {
	c.flagsMutex.Lock()
	defer c.flagsMutex.Unlock()
	rootOrgs, err := GetRootOrgs(c.apiKey, c.Config.ServerUrl, c.Config.FlagConfigPollerRequestTimeout)
	if err != nil {
		return err
	}
	c.flags = rootOrgs
	return nil
}

func (c *Client) Evaluate(orgId any) (any, bool) {
	c.flagsMutex.RLock()
	defer c.flagsMutex.RUnlock()

	orgIdStr, ok := orgId.(string)
	if !ok {
		return nil, false
	}
	orgIdInt, err := strconv.ParseInt(orgIdStr, 10, 64)
	if err != nil {
		return nil, false
	}
	org, ok := (*c.flags)[model.OrgId(orgIdInt)]
	return org, ok
}
