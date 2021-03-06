package taskrunner

import (
	"encoding/json"
	"sync"

	logging "github.com/codemodify/systemkit-logging"

	helpersStrings "github.com/codemodify/systemkit-helpers-conv"
	helpersGuid "github.com/codemodify/systemkit-helpers-guid"
	helpersReflect "github.com/codemodify/systemkit-helpers-reflection"
)

// Preparer - Does the "factory" and sets the `RunTaskInstance`
type Preparer interface {
	Prepare(taskRunner *TaskRunner, runParamsAsBytes []byte)
}

// TaskRunner -
type TaskRunner struct {
	preparer Preparer
}

// NewTaskRunner -
func NewTaskRunner(preparer Preparer) *TaskRunner {
	return &TaskRunner{
		preparer: preparer,
	}
}

// Run -
func (thisRef *TaskRunner) Run(task *Task) {
	thisRef.prep(task)
	thisRef.run(task, nil, task.Label, -1)
}

func (thisRef *TaskRunner) prep(task *Task) {
	if helpersStrings.IsNullOrEmpty(task.ID) {
		task.ID = helpersGuid.NewGUID()
	}

	if !helpersStrings.IsNullOrEmpty(task.Run) {

		// "Behave" like a Factory pattern

		// Get the bytes
		runParamsAsBytes, err := json.Marshal(task.RunParams)
		if err != nil {
			logging.Errorf("%s, %s, %s", err.Error(), task.RunParams, helpersReflect.GetThisFuncName())

			return
		}

		// Expected to set the `RunTaskInstance`
		thisRef.preparer.Prepare(thisRef, runParamsAsBytes)
	}

	for _, t := range task.SeqTasks {
		thisRef.prep(t)
	}

	for _, t := range task.Tasks {
		thisRef.prep(t)
	}
}

func (thisRef *TaskRunner) run(task *Task, wg *sync.WaitGroup, tagPrefix string, callStackLevel int) {
	if task.RunTaskInstance != nil {
		task.RunTaskInstance.Run(tagPrefix+" / "+task.Label, callStackLevel)

		if wg != nil {
			wg.Done()
		}
	} else if len(task.SeqTasks) > 0 {
		for _, t := range task.SeqTasks {
			thisRef.run(t, nil, tagPrefix, callStackLevel+1)
		}

		if wg != nil {
			wg.Done()
		}
	} else if len(task.Tasks) > 0 {
		wg := sync.WaitGroup{}

		for _, t := range task.Tasks {
			wg.Add(1)

			go thisRef.run(t, &wg, tagPrefix, callStackLevel+1)
		}

		wg.Wait()
	}
}
