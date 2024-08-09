package web

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileUploader struct {
	FileField   string
	DstPathFunc func(fh *multipart.FileHeader) string
}

func (f *FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		src, srcHeader, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("upload failed, no data found")
			log.Fatalln(err)
			return
		}
		defer src.Close()
		dst, err := os.OpenFile(f.DstPathFunc(srcHeader), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("upload failed, open file failed")
			log.Fatalln(err)
			return
		}
		defer dst.Close()

		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			ctx.RespStatusCode = 500
			ctx.RespData = []byte("upload failed, copy failed")
			log.Fatalln(err)
			return

		}
		ctx.RespData = []byte("upload success")
	}
}

type FileDownloader struct {
	Dir string
}

func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		req, _ := ctx.QueryValue("file").String()
		path := filepath.Join(f.Dir, filepath.Clean(req))
		fn := filepath.Base(path)
		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}

type StaticResourceHandlerOption func(h *StaticResourceHandler)

type StaticResourceHandler struct {
	dir                     string
	pathPrefix              string
	extensionContentTypeMap map[string]string

	//cache limit
	cache       *lru.Cache
	maxFileSize int
}

type fileCacheItem struct {
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

func NewStaticResourceHandler(dir string, pathPrefix string, options ...StaticResourceHandlerOption) *StaticResourceHandler {
	h := &StaticResourceHandler{
		dir:        dir,
		pathPrefix: pathPrefix,
		extensionContentTypeMap: map[string]string{
			// text
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}
	for _, opt := range options {
		opt(h)
	}
	return h
}

func WithFileCache(maxCacheSize int, maxCacheFileCnt int) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		c, err := lru.New(maxCacheFileCnt)
		if err != nil {
			log.Fatalln("web: create lru cache failed")
		}
		h.maxFileSize = maxCacheSize
		h.cache = c
	}
}

func WithMoreExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		for ext, contentType := range extMap {
			h.extensionContentTypeMap[ext] = contentType
		}
	}
}

func (h *StaticResourceHandler) Handle(ctx *Context) {
	req, _ := ctx.PathValue("file").String()
	if item, ok := h.readFileFromData(req); ok {
		log.Println("web: read file from cache")
		h.writeItemAsResponse(item, ctx.Resp)
		return
	}
}

func (h *StaticResourceHandler) cacheFile(item *fileCacheItem) {
	if h.cache == nil {
		return
	}
	if item.fileSize > h.maxFileSize {
		return
	}
	h.cache.Add(item.fileName, item)
}

func (h *StaticResourceHandler) readFileFromData(fileName string) (*fileCacheItem, bool) {
	if h.cache == nil {
		return nil, false
	}
	if item, ok := h.cache.Get(fileName); ok {
		return item.(*fileCacheItem), true
	}
	return nil, false

}

func (h *StaticResourceHandler) writeItemAsResponse(item *fileCacheItem, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", item.contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", item.fileSize))
	_, _ = w.Write(item.data)
}

func getFileExt(name string) string {
	index := strings.LastIndex(name, ".")
	if index == len(name)-1 {
		return ""
	}
	return name[index+1:]
}
