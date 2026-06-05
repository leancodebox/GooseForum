package optlogger

import "testing"

func TestOptEnumNameAndTargetType(t *testing.T) {
	tests := []struct {
		name       string
		value      OptEnum
		wantName   string
		wantTarget TargetTypeEnum
	}{
		{name: "edit user", value: EditUser, wantName: "操作用户", wantTarget: User},
		{name: "edit article", value: EditArticle, wantName: "编辑文章", wantTarget: Article},
		{name: "unknown", value: OptEnum(99), wantName: "", wantTarget: System},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Name(); got != tt.wantName {
				t.Fatalf("Name() = %q, want %q", got, tt.wantName)
			}
			if got := tt.value.TargetTypeEnum(); got != tt.wantTarget {
				t.Fatalf("TargetTypeEnum() = %v, want %v", got, tt.wantTarget)
			}
		})
	}
}

func TestTargetTypeEnumName(t *testing.T) {
	tests := []struct {
		value TargetTypeEnum
		want  string
	}{
		{value: System, want: "系统"},
		{value: User, want: "用户"},
		{value: DocProject, want: "文档项目"},
		{value: DocVersion, want: "文档版本"},
		{value: DocContent, want: "文档内容"},
		{value: TargetTypeEnum(99), want: ""},
	}

	for _, tt := range tests {
		if got := tt.value.Name(); got != tt.want {
			t.Fatalf("Name() = %q, want %q", got, tt.want)
		}
	}
}

func TestEnumToInt(t *testing.T) {
	if got := EditArticle.toInt(); got != 1 {
		t.Fatalf("OptEnum.toInt() = %d, want 1", got)
	}
	if got := TargetTypeEnum(DocContent).toInt(); got != 5 {
		t.Fatalf("TargetTypeEnum.toInt() = %d, want 5", got)
	}
}
