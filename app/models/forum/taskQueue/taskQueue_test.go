package taskQueue

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestEntityJSONDoesNotExposeTaskPayload(t *testing.T) {
	encoded, err := json.Marshal(Entity{Id: 1, Type: "activation", TaskJson: `{"token":"secret"}`})
	if err != nil {
		t.Fatalf("marshal task entity: %v", err)
	}
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "taskJson") {
		t.Fatalf("task payload leaked through JSON: %s", encoded)
	}
}

func TestClaimOnlyMovesPendingTaskOnce(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate task queue: %v", err)
	}
	typeValue := "claim-test-" + time.Now().Format("150405.000000000")
	task := &Entity{Type: typeValue, Status: StatusPending, TaskJson: `{}`}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	t.Cleanup(func() { db.Delete(&Entity{}, task.Id) })

	claimed, err := Claim(task.Id)
	if err != nil || !claimed {
		t.Fatalf("first Claim() = claimed %v, err %v; want true, nil", claimed, err)
	}
	claimed, err = Claim(task.Id)
	if err != nil || claimed {
		t.Fatalf("second Claim() = claimed %v, err %v; want false, nil", claimed, err)
	}
}

func TestRetryOnlyRequeuesFailedMatchingTask(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate task queue: %v", err)
	}
	task := &Entity{Type: "registration_invitation", Status: StatusFailed, RetryCount: 3, TaskJson: `{}`, LastError: "smtp"}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create failed task: %v", err)
	}
	t.Cleanup(func() { db.Delete(&Entity{}, task.Id) })
	retried, err := Retry(task.Id, "registration_invitation")
	if err != nil || !retried {
		t.Fatalf("Retry() = %v, %v; want true, nil", retried, err)
	}
	var got Entity
	if err := db.First(&got, task.Id).Error; err != nil {
		t.Fatalf("reload task: %v", err)
	}
	if got.Status != StatusPending || got.RetryCount != 0 || got.LastError != "" {
		t.Fatalf("retried task = %+v", got)
	}
	if retried, err := Retry(task.Id, "registration_invitation"); err != nil || retried {
		t.Fatalf("second Retry() = %v, %v; want false, nil", retried, err)
	}
}
