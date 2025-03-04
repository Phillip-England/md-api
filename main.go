package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/Phillip-England/vbf"
	"github.com/PuerkitoBio/goquery"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

func main() {

	mux, gCtx := vbf.VeryBestFramework()

	vbf.HandleStaticFiles(mux)
	vbf.HandleFavicon(mux)

	vbf.AddRoute("POST /", mux, gCtx, func(w http.ResponseWriter, r *http.Request) {
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = "dracula"
		}
		const maxBodySize = 1 * 1024 * 1024 // 1MB
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		defer r.Body.Close()
		markdownContent := string(body)
		md := goldmark.New(
			goldmark.WithExtensions(
				highlighting.NewHighlighting(
					highlighting.WithStyle(theme),
					highlighting.WithFormatOptions(
						chromahtml.WithLineNumbers(true),
					),
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				goldmarkhtml.WithHardWraps(),
				goldmarkhtml.WithXHTML(),
				goldmarkhtml.WithUnsafe(),
			),
		)
		var buf bytes.Buffer
		err = md.Convert([]byte(markdownContent), &buf)
		if err != nil {
			vbf.WriteJSON(w, 200, map[string]interface{}{
				"message": "failed to parse markdown content",
			})
			return
		}
		str := buf.String()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
		if err != nil {
			vbf.WriteJSON(w, 200, map[string]interface{}{
				"message": "failed to parse markdown content",
			})
			return
		}
		modifiedHTML, err := doc.Html()
		if err != nil {
			vbf.WriteJSON(w, 200, map[string]interface{}{
				"message": "failed to parse markdown content",
			})
			return
		}
		vbf.WriteJSON(w, 200, map[string]interface{}{
			"message": "Markdown received",
			"content": modifiedHTML,
		})
	}, vbf.MwCORS, vbf.MwLogger)

	err := vbf.Serve(mux, "8080")
	if err != nil {
		panic(err)
	}

}
