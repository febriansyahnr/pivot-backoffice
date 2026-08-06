package conductor_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/conductor"

	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/conductor-sdk/conductor-go/sdk/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleClient(t *testing.T) {
	if os.Getenv("TEST_CONDUCTOR") != "true" {
		t.Skip("Skip testing conductor workflow")
	}

	const (
		baseURL  = "http://localhost:5000/conductor/api" // NOSONAR
		username = "admin"                               // NOSONAR
		password = "qwerty"                              // NOSONAR
	)

	t.Run("Invalid Address", func(t *testing.T) {
		_, err := NewClient(Config{
			BaseURL: "http://localhost:65000", // NOSONAR
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "connect: connection refused")
	})

	t.Run("Invalid Authentication", func(t *testing.T) {
		_, err := NewClient(Config{
			BaseURL: baseURL,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "401 Authorization Required")
	})

	client, err := NewClient(Config{
		BaseURL: baseURL,
		Authentication: &BasicAuthentication{
			Username: username,
			Password: password,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, client)

	worker := client.Workers()
	defer worker.Close(context.Background())

	go func() {
		worker.RunWorkers(t.Context(), []WorkerDefinition{
			{
				TaskName: "greet",
				Handler: func(_ context.Context, task *Task) (*TaskResult, error) {
					name, _ := task.InputData["name"].(string)
					assert.Equal(t, "Gopher", name)

					taskResult := model.NewTaskResultFromTask(task)
					taskResult.OutputData = map[string]any{
						"greetings": "Hi " + name,
					}
					taskResult.Status = model.CompletedTask

					return taskResult, nil
				},
				BatchSize:    1,
				PollInterval: 100 * time.Millisecond,
			},
		})
	}()

	executor := client.WorkflowExecutor()

	log.Println("Start workflow definition")

	wf := workflow.NewConductorWorkflow(executor).
		Name("greetings").
		Version(1).
		Description("Greetings workflow - Greets a user by their name").
		TimeoutPolicy(workflow.TimeOutWorkflow, 600).
		OwnerEmail("testing@service.local")

	greet := workflow.NewSimpleTask("greet", "greet_ref").
		Input("name", "${workflow.input.name}")

	wf.Add(greet)

	wf.OutputParameters(map[string]any{
		"greetings": greet.OutputRef("greetings"),
	})

	log.Println("Register workflow")

	if err := wf.Register(true); err != nil {
		require.Contains(t, err.Error(), "Workflow with greetings.1 already exists!")
		log.Println("Workflow already exists, proceed to run it")
	}

	id, err := executor.StartWorkflowWithContext(t.Context(), &model.StartWorkflowRequest{
		Name:     "greetings",
		Version:  1,
		Priority: 1,
		Input: map[string]string{
			"name": "Gopher",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	log.Println("Started workflow id", id)

	run, err := executor.MonitorExecution(id)
	require.NoError(t, err)

	select {
	case channel := <-run:
		greetings, _ := channel.Output["greetings"].(string)
		assert.Equal(t, "Hi Gopher", greetings)

	case <-time.After(3 * time.Second):
		t.Log("Waiting timeout exceeded")
		t.Fail()
	}
}
