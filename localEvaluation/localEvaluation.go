package localEvaluation

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/experiment"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/experiment/local"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/rootOrg"
)

var (
	client                                    *local.Client
	rootOrgClient                             *rootOrg.Client
	localEvaluationConfigDebug                = true
	localEvaluationConfigServerUrl            = "https://api.lambdatest.com"
	localEvaluationConfigPollInterval         = 120
	localEvaluationConfigPollerRequestTimeout = 60
	localEvaluationDeploymentKey              = "server-jAqqJaX3l8PgNiJpcv9j20ywPzANQQFh"
	retries                                   = 5
)

type Variant struct {
	Value    string                 `json:"value,omitempty"`
	Payload  interface{}            `json:"payload,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UserProperties struct {
	OrgId            string `json:"org_id,omitempty"`
	UserId           string `json:"user_id,omitempty"`
	OrgName          string `json:"org_name,omitempty"`
	Username         string `json:"username,omitempty"`
	UserStatus       string `json:"user_status,omitempty"`
	Email            string `json:"email,omitempty"`
	Plan             string `json:"plan,omitempty"`
	SubscriptionType string `json:"subscription_type,omitempty"`
	HubRegion        string `json:"hub_region,omitempty"`
	InfraProvider    string `json:"infra_provider,omitempty"`
	TemplateId       string `json:"template_id,omitempty"`
}

func initVars() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("No .env file found")
	} else {
		fmt.Printf(".env file loaded")
	}

	if os.Getenv("LOCAL_EVALUATION_CONFIG_DEBUG") != "" {
		localEvaluationConfigDebug, _ = strconv.ParseBool(os.Getenv("LOCAL_EVALUATION_CONFIG_DEBUG"))
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_SERVER_URL") != "" {
		localEvaluationConfigServerUrl = os.Getenv("LOCAL_EVALUATION_CONFIG_SERVER_URL")
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_POLL_INTERVAL") != "" {
		localEvaluationConfigPollInterval, _ = strconv.Atoi(os.Getenv("LOCAL_EVALUATION_CONFIG_POLL_INTERVAL"))
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_POLLER_REQUEST_TIMEOUT") != "" {
		localEvaluationConfigPollerRequestTimeout, _ = strconv.Atoi(os.Getenv("LOCAL_EVALUATION_CONFIG_POLLER_REQUEST_TIMEOUT"))
	}
	if os.Getenv("LOCAL_EVALUATION_DEPLOYMENT_KEY") != "" {
		localEvaluationDeploymentKey = os.Getenv("LOCAL_EVALUATION_DEPLOYMENT_KEY")
	}
}

func Initialize() {
	initVars()
	config := local.Config{
		Debug:                          localEvaluationConfigDebug,
		ServerUrl:                      localEvaluationConfigServerUrl,
		FlagConfigPollerInterval:       time.Duration(localEvaluationConfigPollInterval) * time.Second,
		FlagConfigPollerRequestTimeout: time.Duration(localEvaluationConfigPollerRequestTimeout) * time.Second,
	}
	client = local.Initialize(localEvaluationDeploymentKey, &config)
	var err error
	for i := 0; i < retries; i++ {
		err = client.Start()
		if err != nil {
			err = fmt.Errorf("unable to create local evaluation client with given config %+v attempt:%v with error %s", config, i+1, err.Error())
			continue
		} else {
			InitializeRootOrg()
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %+v with error %s", config, err.Error())
		panic(err)
	}
}

func InitializeWithConfig(conf local.Config, deploymentKey string) {
	client = local.Initialize(deploymentKey, &conf)
	var err error
	for i := 0; i < retries; i++ {
		err = client.Start()
		if err != nil {
			err = fmt.Errorf("unable to create local evaluation client with given config %+v attempt:%v with error %s", conf, i+1, err.Error())
			continue
		} else {
			InitializeRootOrg()
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %+v with error %s", conf, err.Error())
		panic(err)
	}
}

func InitializeRootOrg() error {
	rootOrgClient = rootOrg.NewClient(localEvaluationDeploymentKey, &rootOrg.Config{
		ServerUrl:                      localEvaluationConfigServerUrl,
		FlagConfigPollerInterval:       time.Duration(localEvaluationConfigPollInterval) * time.Second,
		FlagConfigPollerRequestTimeout: time.Duration(localEvaluationConfigPollerRequestTimeout) * time.Second,
	})
	var err error
	for i := 0; i < retries; i++ {
		err = rootOrgClient.Start()
		if err != nil {
			err = fmt.Errorf("unable to get root orgs with given config %+v attempt:%v with error %s", rootOrgClient.Config, i+1, err.Error())
			continue
		} else {
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("unable to get root orgs with given config %+v with error %s", rootOrgClient.Config, err.Error())
		return err
	}
	return nil
}

func fetch(user UserProperties) (map[string]experiment.Variant, error) {
	userProp := map[string]any{
		"org_id":            user.OrgId,
		"org_name":          user.OrgName,
		"username":          user.Username,
		"email":             user.Email,
		"plan":              user.Plan,
		"subscription_type": user.SubscriptionType,
		"hub_region":        user.HubRegion,
		"infra_provider":    user.InfraProvider,
		"template_id":       user.TemplateId,
	}
	expUser := experiment.User{
		UserId:         user.UserId,
		UserProperties: userProp,
	}

	// Evaluate root org to get the parent org id
	if expUser.UserProperties["org_id"] != "" && rootOrgClient != nil {
		rootOrg, ok := rootOrgClient.Evaluate(expUser.UserProperties["org_id"])
		if ok {
			expUser.UserProperties["org_id"] = rootOrg
		}
	}

	result, err := client.EvaluateV2(&expUser, []string{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getValue(flagName string, user UserProperties) Variant {
	result, _ := fetch(user)
	if len(result) != 0 {
		if value, ok := result[flagName]; ok {
			return Variant{
				Key:     value.Key,
				Payload: value.Payload,
			}
		}
	}
	return Variant{}
}

func getMapOfValue(user UserProperties) map[string]interface{} {
	flags := make(map[string]interface{})
	result, _ := fetch(user)
	if len(result) != 0 {
		for k, v := range result {
			if v.Value != "" {
				flags[k] = v.Value
			}
		}
	}
	return flags
}

func GetFeatureFlagString(flagName string, user UserProperties) string {
	data := getValue(flagName, user)
	return data.Key
}

func GetFeatureFlagBool(flagName string, user UserProperties) bool {
	data := getValue(flagName, user)
	if val, err := strconv.ParseBool(data.Key); err == nil {
		return val
	}
	return false
}

func GetFeatureFlagPayload(flagName string, user UserProperties) map[string]interface{} {
	data := getValue(flagName, user)
	mapData := make(map[string]interface{})
	mapData["value"] = data.Key
	mapData["payload"] = data.Payload
	return mapData
}

func GetFeatureFlagByOrg(user UserProperties) map[string]interface{} {
	data := getMapOfValue(user)
	return data
}

func GetFeatureFlagAllDataByOrg(user UserProperties) map[string]experiment.Variant {
	result, _ := fetch(user)
	return result
}
