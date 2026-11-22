package imageuploadservice

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
)

func TestSharedAvatarThumbnail(t *testing.T) {
	if err := db4fileconnect.Connect().AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := dbconnect.Connect().AutoMigrate(&fileUsage.Entity{}); err != nil {
		t.Fatal(err)
	}
	src := image.NewNRGBA(image.Rect(0, 0, 800, 400))
	for y := range 400 {
		for x := 200; x < 600; x++ {
			src.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	original, err := Save(ctx, 7, buf.Bytes(), "original.png", true)
	if err != nil {
		t.Fatal(err)
	}
	key, err := DefaultAvatar(ctx, original.Name)
	if err != nil {
		t.Fatal(err)
	}
	again, err := DefaultAvatar(ctx, original.Name)
	if err != nil || again != key || key == original.Name {
		t.Fatalf("shared thumbnail %q, %q: %v", key, again, err)
	}
	valid, err := fileUsage.AllAdminUploads([]string{original.Name, key})
	if err != nil || !valid {
		t.Fatalf("admin references: %v, %v", valid, err)
	}
	valid, err = fileUsage.AllAdminUploads([]string{key, "unknown"})
	if err != nil || valid {
		t.Fatalf("missing reference: %v, %v", valid, err)
	}
	obj, err := filestorage.Open(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	decoded, err := png.Decode(obj.Body)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds() != image.Rect(0, 0, 256, 256) {
		t.Fatalf("size = %v", decoded.Bounds())
	}
	if _, g, _, a := decoded.At(0, 0).RGBA(); g != 65535 || a != 65535 {
		t.Fatal("thumbnail was not center cropped")
	}
}
