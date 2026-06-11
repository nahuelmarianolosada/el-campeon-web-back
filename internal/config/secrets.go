package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// LoadSecretsFromSSM fetches every parameter under pathPrefix from AWS SSM
// Parameter Store (with SecureString decryption) and exports each one as an
// environment variable. A parameter at "/el-campeon/prod/DB_PASSWORD" is
// exported as DB_PASSWORD, so the rest of Load() can keep reading os.Getenv
// without knowing where the value came from. SSM is authoritative: if a
// parameter exists in SSM, it overrides any pre-existing env var.
//
// IAM permissions required on the task/instance role:
//   - ssm:GetParametersByPath on arn:aws:ssm:<region>:<account>:parameter<pathPrefix>*
//   - kms:Decrypt on the KMS key used for the SecureString parameters
//     (the AWS-managed alias/aws/ssm is enough if you didn't bring your own key).
func LoadSecretsFromSSM(ctx context.Context, pathPrefix string) error {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}
	client := ssm.NewFromConfig(awsCfg)

	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	var nextToken *string
	loaded := 0
	for {
		out, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(pathPrefix),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		})
		if err != nil {
			return fmt.Errorf("ssm GetParametersByPath %q: %w", pathPrefix, err)
		}

		for _, p := range out.Parameters {
			key := strings.TrimPrefix(aws.ToString(p.Name), pathPrefix)
			// Flatten any nested path: "/el-campeon/prod/db/password" -> "DB_PASSWORD".
			key = strings.ToUpper(strings.ReplaceAll(key, "/", "_"))
			if key == "" {
				continue
			}
			if err := os.Setenv(key, aws.ToString(p.Value)); err != nil {
				return fmt.Errorf("setenv %q: %w", key, err)
			}
			loaded++
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	log.Printf("Loaded %d secrets from SSM path %q", loaded, pathPrefix)
	return nil
}
