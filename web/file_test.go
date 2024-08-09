package web

import (
	"bytes"
	"html/template"
	"mime/multipart"
	"path"
	"testing"
)

func TestFileUploader_Handle(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/upload_page", func(ctx *Context) {
		tpl := template.New("upload")
		tpl, err := tpl.Parse(`
<html>
<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		 <input type="file" name="myfile" />
		 <button type="submit">上传</button>
	</form>
</body>
<html>
`)
		if err != nil {
			t.Fatal(err)
		}

		page := &bytes.Buffer{}
		err = tpl.Execute(page, nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx.RespStatusCode = 200
		ctx.RespData = page.Bytes()
	})
	s.Post("/upload", (&FileUploader{
		// field name in the template (tag attribute name)
		FileField: "myfile",
		DstPathFunc: func(fh *multipart.FileHeader) string {
			return path.Join("testdata", "upload", fh.Filename)
		},
	}).Handle())
	s.Start(":8081")
}

func TestFileDownloader_Handle(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/download", (&FileDownloader{

		Dir: "./testdata/download",
	}).Handle())
	//  localhost:8081/download?file=test.txt
	s.Start(":8081")
}

func TestStaticResourceHandler_Handle(t *testing.T) {
	s := NewHTTPServer()
	handler := NewStaticResourceHandler("./testdata/img", "/img")
	s.Get("/img/:file", handler.Handle)
	// localhost:8081/img/come_on_baby.jpg
	s.Start(":8081")
}
