package eventbus

import (
	"os/exec"
	"strings"
	"testing"
)

func TestEventbusDoesNotImportEventhandlers(t *testing.T) {
	output, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed: %v\n%s", err, output)
	}
	if strings.Contains(string(output), "/app/service/eventhandlers") {
		t.Fatalf("eventbus must not import service eventhandlers:\n%s", output)
	}
}
