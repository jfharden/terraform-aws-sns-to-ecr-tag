package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	awsSdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	AwsRegion        string
	RandomName       string
	TerraformOptions *terraform.Options
	WorkingDir       string
}

func LoadTestData(t *testing.T, workingDir string) *TestData {
	return &TestData{
		AwsRegion:        test_structure.LoadString(t, workingDir, "awsRegion"),
		RandomName:       test_structure.LoadString(t, workingDir, "randomName"),
		TerraformOptions: test_structure.LoadTerraformOptions(t, workingDir),
		WorkingDir:       workingDir,
	}
}

func (testData *TestData) Save(t *testing.T) {
	test_structure.SaveString(t, testData.WorkingDir, "randomName", testData.RandomName)
	test_structure.SaveString(t, testData.WorkingDir, "awsRegion", testData.AwsRegion)
	test_structure.SaveTerraformOptions(t, testData.WorkingDir, testData.TerraformOptions)
}

func Setup(t *testing.T, workingDir string) *TestData {
	randomName := fmt.Sprintf("ecs-test-%s", strings.ToLower(random.UniqueId()))
	awsRegion := "eu-west-1"

	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/simple/",

		Vars: map[string]interface{}{
			"name": randomName,
			"tags": map[string]string{
				"TerraformTest": randomName,
			},
		},

		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}

	testData := &TestData{
		AwsRegion:        awsRegion,
		RandomName:       randomName,
		TerraformOptions: terraformOptions,
		WorkingDir:       workingDir,
	}

	testData.Save(t)

	return testData
}

func TestSNSToECR(t *testing.T) {
	workingDir := filepath.Join(".terratest-working-dir", t.Name())

	defer test_structure.RunTestStage(t, "destroy", func() {
		testData := LoadTestData(t, workingDir)
		terraform.Destroy(t, testData.TerraformOptions)
	})

	test_structure.RunTestStage(t, "init_apply", func() {
		testData := Setup(t, workingDir)

		terraform.InitAndApply(t, testData.TerraformOptions)

		pushImage(t, testData)
	})

	validators := map[string]func(*testing.T, *TestData){
		"ValidateTagging": ValidateTagging,
	}

	test_structure.RunTestStage(t, "validate", func() {
		testData := LoadTestData(t, workingDir)

		t.Run("ParalellTests", func(t *testing.T) {
			for name, validator := range validators {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					validator(t, testData)
				})
			}
		})
	})
}

func pushImage(t *testing.T, testData *TestData) {
	shell.RunCommand(t, shell.Command{
		Command: "../examples/shared/push-image.sh",
		Args: []string{
			aws.GetAccountId(t),
			testData.AwsRegion,
			testData.RandomName,
			"latest",
		},
	})
}

type SNSPayload struct {
	RepoName    string `json:"ecr_repo_name"`
	TagToUpdate string `json:"ecr_tag_to_update"`
	TagToAdd    string `json:"ecr_tag_to_add"`
}

func ValidateTagging(t *testing.T, testData *TestData) {
	awsSession := session.Must(session.NewSession(&awsSdk.Config{
		Region: awsSdk.String(testData.AwsRegion),
	}))

	expectedDigest, err := GetInitialImageDigest(t, testData, awsSession)
	if err != nil {
		t.Fatalf("Coundln't get initial image digest: %s", err)
		return
	}

	payload := &SNSPayload{
		RepoName:    testData.RandomName,
		TagToUpdate: "latest",
		TagToAdd:    random.UniqueId(),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Coundln't encode json payload: %s", err)
		return
	}

	snsService := sns.New(awsSession)

	_, err = snsService.Publish(&sns.PublishInput{
		TopicArn: awsSdk.String(terraform.Output(t, testData.TerraformOptions, "sns_topic_arn")),
		Message:  awsSdk.String(string(jsonPayload)),
	})
	if err != nil {
		t.Fatalf("Coundln't publish message to SNS topic: %s", err)
		return
	}

	actualDigest, err := GetImageDigestForTag(t, payload.TagToAdd, testData, awsSession, 12)
	if err != nil {
		t.Fatalf("Coundln't get image with tag %s: %s", payload.TagToAdd, err)
		return
	}

	assert.Equal(t, expectedDigest, actualDigest, "Digest of tagged image does not match image expected to be tagged")
}

func GetInitialImageDigest(t *testing.T, testData *TestData, awsSession *session.Session) (string, error) {
	digest, err := GetImageDigestForTag(t, "latest", testData, awsSession, 0)
	return digest, err
}

func GetImageDigestForTag(t *testing.T, tag string, testData *TestData, awsSession *session.Session, retries int) (string, error) {
	ecrService := ecr.New(awsSession)

	digest, err := retry.DoWithRetryE(t, fmt.Sprintf("Get image digest for tag %s", tag), retries, 5*time.Second, func() (string, error) {
		listOutput, err := ecrService.ListImages(&ecr.ListImagesInput{
			RepositoryName: awsSdk.String(testData.RandomName),
		})
		if err != nil {
			return "", err
		}

		for _, imageID := range listOutput.ImageIds {
			if *imageID.ImageTag == tag {
				return *imageID.ImageDigest, nil
			}
		}

		return "", fmt.Errorf("Image Tag %s not found in repository %s", tag, testData.RandomName)
	})

	return digest, err
}
