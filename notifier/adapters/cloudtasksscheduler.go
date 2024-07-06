package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/4Kaze/birthdaybot/common"
	"github.com/4Kaze/birthdaybot/notifier/core"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CloudTasksScheduler struct {
	client        *cloudtasks.Client
	clock         core.Clock
	queuePath     string
	taskDeadlline time.Duration
	taskDelay     time.Duration
}

func NewCloudTasksScheduler(client *cloudtasks.Client, clock core.Clock, queuePath string, taskDeadline time.Duration, taskDelay time.Duration) *CloudTasksScheduler {
	return &CloudTasksScheduler{
		client:        client,
		clock:         clock,
		queuePath:     queuePath,
		taskDeadlline: taskDeadline,
		taskDelay:     taskDelay,
	}
}

func (scheduler *CloudTasksScheduler) Schedule(ctx context.Context, birthday core.Birthday, serviceUrl string) {
	birthdayJson, err := json.Marshal(common.BirthdayJson{
		ChatId: birthday.ChatId,
		UserId: birthday.UserId,
		Name:   birthday.Name,
	})
	if err != nil {
		common.ErrorLogger.Printf("Could not marshal birthday: %v to json, due to: %v\n", birthday, err)
		return
	}
	taskName := fmt.Sprintf("%s/tasks/%v%v%v", scheduler.queuePath, birthday.ChatId, birthday.UserId, scheduler.clock.Now().YearDay())
	req := &taskspb.CreateTaskRequest{
		Parent: scheduler.queuePath,
		Task: &taskspb.Task{
			ScheduleTime:     timestamppb.New(scheduler.clock.Now().Add(time.Second * 10)),
			DispatchDeadline: durationpb.New(scheduler.taskDeadlline),
			Name:             taskName,
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        serviceUrl,
					Body:       birthdayJson,
				},
			},
		},
	}
	task, err := scheduler.client.CreateTask(ctx, req)
	if err != nil {
		common.ErrorLogger.Printf("Could not create birthday task: %v (%s), due to: %v\n", string(birthdayJson), taskName, err)
		return
	}
	log.Printf("Created a task %s for birthday: %v\n", task, string(birthdayJson))
}
