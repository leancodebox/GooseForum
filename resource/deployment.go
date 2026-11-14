package resource

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/buildinfo"
)

// DeploymentID is shared by the initial HTML and status endpoint. It is stable
// across replicas of the same build and includes the frontend asset manifest.
var DeploymentID = sync.OnceValue(func() string {
	info := buildinfo.Get()
	manifest, _ := fs.ReadFile(GetTemplateFS(), "static/dist/react/.vite/manifest.json")
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s\n%s\n%s\n%s", info.Version, info.Commit, info.BuildDate, manifest))))
})
