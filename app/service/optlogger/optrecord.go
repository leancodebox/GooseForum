package optlogger

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/spf13/cast"
)

func UserOpt(userId uint64, optType OptEnum, targetId any, msg string) {
	entity := optRecord.Entity{
		OptUserId: userId, OptType: optType.toInt(), TargetType: optType.TargetTypeEnum().toInt(),
		TargetId: cast.ToString(targetId), OptInfo: msg, CreatedAt: time.Now()}
	optRecord.Create(&entity)
}

type MessageParams map[string]any

type MessagePayload struct {
	MessageCode string        `json:"messageCode"`
	Params      MessageParams `json:"params,omitempty"`
}

func UserOptCode(userId uint64, optType OptEnum, targetId any, messageCode string, params MessageParams) {
	payload := MessagePayload{
		MessageCode: messageCode,
		Params:      params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal operation log payload", "messageCode", messageCode, "err", err)
		UserOpt(userId, optType, targetId, messageCode)
		return
	}
	UserOpt(userId, optType, targetId, string(data))
}
