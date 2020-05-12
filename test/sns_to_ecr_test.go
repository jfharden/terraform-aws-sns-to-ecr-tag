package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	// awsSdk "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/ecr"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
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
		"ValidateTaggingWorks": ValidateTaggingWorks,
	}

	test_structure.RunTestStage(t, "validate", func() {
		testData := LoadTestData(t, workingDir)

		t.Run("ParalellTests", func(t *testing.T) {
			for name, validator := range validators {
				t.Run(name, func(t *testing.T) {
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

func ValidateTaggingWorks(t *testing.T, testData *TestData) {
	assert.True(t, false)
}
