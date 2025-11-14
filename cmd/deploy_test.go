package cmd

import (
	"testing"

	"github.com/friendsofshopware/shopmon-cli/internal/deployment"
	"github.com/stretchr/testify/assert"
)

func TestDeployCommandRequiresToken(t *testing.T) {
	t.Setenv("SHOPMON_DEPLOY_TOKEN", "")

	originalFactory := newDeploymentService
	defer func() { newDeploymentService = originalFactory }()

	newDeploymentService = func() *deployment.Service {
		t.Fatalf("deployment service should not be created without SHOPMON_DEPLOY_TOKEN")
		return nil
	}

	err := deployCmd.RunE(deployCmd, []string{"deploy", "--", "echo", "test"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SHOPMON_DEPLOY_TOKEN")
}

func TestDeployCommandRunsServiceWhenTokenSet(t *testing.T) {
	t.Setenv("SHOPMON_DEPLOY_TOKEN", "token")

	originalFactory := newDeploymentService
	originalRunner := runDeploymentService
	defer func() {
		newDeploymentService = originalFactory
		runDeploymentService = originalRunner
	}()

	newDeploymentService = func() *deployment.Service {
		return &deployment.Service{}
	}

	var runCalled bool
	runDeploymentService = func(service *deployment.Service, args []string) error {
		runCalled = true
		assert.Equal(t, []string{"deploy", "--", "echo", "test"}, args)
		return nil
	}

	err := deployCmd.RunE(deployCmd, []string{"deploy", "--", "echo", "test"})

	assert.NoError(t, err)
	assert.True(t, runCalled, "expected deployment service to run")
}
