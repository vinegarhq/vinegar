package main

import (
	"embed"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

//go:embed resources/*.ui resources/logo.png
var resources embed.FS

const style = `
.routine {
	font-weight: 800;
	font-size: 141%;
}
`

func resource(name string) string {
	b, err := resources.ReadFile("resources/" + name)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func setLogoImage(image *gtk.Image) {
	// Workaround for gdkbixbuf.NewPixbufFromData being non-functional
	logoData := []byte(resource("logo.png"))

	l := gdkpixbuf.NewPixbufLoader()
	l.Write(uintptr(unsafe.Pointer(&logoData[0])), uint(len(logoData)))
	l.Close()
	pb := l.GetPixbuf()
	l.Unref()
	image.SetFromPixbuf(pb)
	pb.Unref()
}
