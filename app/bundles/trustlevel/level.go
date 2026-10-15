package trustlevel

type Level int16

const (
	Newcomer  Level = 0
	Member    Level = 100
	Regular   Level = 200
	Core      Level = 1000
	MaxTopics       = 20
)

type Definition struct {
	Level     Level  `json:"level"`
	Name      string `json:"name"`
	SystemKey string `json:"systemKey,omitempty"`
}

func Definitions() []Definition {
	return []Definition{{Newcomer, "新用户", ""}, {Member, "成员", "trust:100"}, {Regular, "稳定成员", "trust:200"}, {Core, "社区骨干", "trust:1000"}}
}

func Valid(level Level) bool {
	return level == Newcomer || level == Member || level == Regular || level == Core
}

func Name(level Level) string {
	for _, d := range Definitions() {
		if d.Level == level {
			return d.Name
		}
	}
	return ""
}

func SystemKeys() []string { return []string{"trust:100", "trust:200", "trust:1000"} }
