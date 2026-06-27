package loader

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/caches"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/times"
	"warm-nest/pkg/utils/transforms"
)

// LoadSecretTo loads a secret from AWS Secrets Manager and maps its keys to the fields of the provided struct based on struct tags.
func LoadSecretTo[V any](secretId string, secretKey string, v *V) error {
	secretMap, err := LoadSecrets(secretId)
	if err != nil {
		return err
	}
	if secretMap == nil {
		return nil
	}
	secretKeys := strings.Split(secretKey, ",")
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		key := field.Tag.Get("secret")
		if key == "" {
			key = field.Tag.Get("json")
		}
		if key == "" {
			continue
		}
		item, ok := slices.FindOne(secretKeys, func(item string) bool {
			_, ok := secretMap[fmt.Sprintf("%s.%s", item, key)]
			return ok
		})
		if !ok {
			continue
		}

		val, _ := secretMap[fmt.Sprintf("%s.%s", item, key)]
		fv := rv.Field(i)

		err = reflects.SetFieldValue(fv, val)
		if err != nil {
			logrus.WithError(err).Errorf("Set secret field value failed! secretId=%s, key=%s, val=%s", secretId, key, val)
			continue
		}
	}
	return nil
}

// LoadSecret loads a specific key from a secret in AWS Secrets Manager.
func LoadSecret(secretId string, secretKey string) (string, bool) {
	secretMap, err := LoadSecrets(secretId)
	if err != nil {
		return "", false
	}
	if secretMap == nil || len(secretMap) == 0 {
		return "", false
	}
	v, ok := secretMap[secretKey]
	return v, ok
}

// LoadSecretByPattern loads a secret from AWS Secrets Manager and filters its keys based on provided regex patterns.
func LoadSecretByPattern(secretId string, patterns ...string) (map[string]string, error) {
	secretMap, err := LoadSecrets(secretId)
	if err != nil {
		return nil, err
	}
	if secretMap == nil || len(patterns) == 0 {
		return secretMap, nil
	}

	filtered := make(map[string]string)
	for k, v := range secretMap {
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			ok, _ := regexp.MatchString(p, k)
			if ok {
				filtered[k] = v
				break
			}
		}
	}
	return filtered, nil
}

// LoadSecrets loads and parses a secret from AWS Secrets Manager given its secretId.
func LoadSecrets(secretId string) (map[string]string, error) {
	if secretId == "" {
		logrus.Warnf("SecretId is empty, please check the secretConfig!")
		return nil, nil
	}
	logrus.Infof("Loading secrets from %s begin...", secretId)
	cacheKey := caches.CacheKey("LoadSecrets", secretId)
	return caches.GetOrLoad[map[string]string](cacheKey, times.FiveMinuteInSec, func() (interface{}, error) {
		secretContent, err := getSecret(secretId)
		if err != nil {
			logrus.WithError(err).Errorf("Get secret failed! sercretId=%s", secretId)
			return nil, fmt.Errorf("get secret %s failed: %w", secretId, err)
		}
		secretConfig := transforms.Unmarshal[map[string]string](secretContent)
		logrus.Infof("Loading secrets from %s succeed", secretId)
		return secretConfig, nil
	})
}

// getSecret retrieves the secret value from AWS Secrets Manager using web identity federation.
func getSecret(secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	client := sts.NewFromConfig(cfg)
	sess := strconv.FormatInt(time.Now().UnixNano(), 10)
	appCreds := aws.NewCredentialsCache(stscreds.NewWebIdentityRoleProvider(
		client,
		os.Getenv("AWS_ROLE_ARN"),
		stscreds.IdentityTokenFile(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")),
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = sess
		}))
	value, err := appCreds.Retrieve(context.TODO())
	if err != nil {
		return "", err
	}
	secretsManager := secretsmanager.New(
		secretsmanager.Options{
			Credentials: credentials.NewStaticCredentialsProvider(value.AccessKeyID, value.SecretAccessKey, value.SessionToken),
			Region:      os.Getenv("AWS_REGION"),
		},
	)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := secretsManager.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", err
	}
	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			return "", err
		}
		secretString = string(decodedBinarySecretBytes[:len])
	}

	return secretString, nil
}
