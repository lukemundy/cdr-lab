package consumerpages

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	"github.com/lukemundy/cdr-lab/pkg/platform/asset"
)

// The data holder image copies the consumer pages build of web/packages/ui into static, so a plain Go build gives
// pages with no styles.
//
//go:embed all:static
var static embed.FS

// staticTypes are the types of the files of the build. The static path serves no other file.
var staticTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".woff2": "font/woff2",
	".txt":   "text/plain; charset=utf-8",
}

func Static() fs.FS {
	files, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	return files
}

// RegisterStatic adds the stylesheet and the fonts of the pages at /static/. They contain no CDR data, so they need no
// authentication. It panics when a file cannot be read, because the files are part of the binary.
func RegisterStatic(mux *http.ServeMux, files fs.FS) {
	err := fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		contentType, found := staticTypes[path.Ext(name)]
		if entry.IsDir() || !found {
			return nil
		}

		data, err := fs.ReadFile(files, name)
		if err != nil {
			return err
		}

		mux.Handle("GET /static/"+name, asset.Handler(contentType, data, ""))

		return nil
	})
	if err != nil {
		panic(err)
	}
}
