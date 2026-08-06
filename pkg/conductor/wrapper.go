package conductor

import "github.com/conductor-sdk/conductor-go/sdk/model"

func NewTaskResultWithNonRetryableError(task *Task, err error) *TaskResult {
	return model.NewTaskResultFromTaskWithError(task, model.NewNonRetryableError(err))
}

func NewTaskResultWithError(task *Task, err error) *TaskResult {
	return model.NewTaskResultFromTaskWithError(task, err)
}

func NewTaskResultWithOutputAndError(task *Task, output any, err error) *TaskResult {
	taskResult := model.NewTaskResultFromTaskWithError(task, err)

	switch val := output.(type) {
	case map[string]any:
		taskResult.OutputData = val

	default:
		taskResult.OutputData, _ = model.ConvertToMap(output)
	}
	return taskResult
}

func NewTaskResult(task *Task) *TaskResult {
	return model.NewTaskResultFromTask(task)
}
