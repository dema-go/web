package msgo

import (
	"errors"
	"github.com/demo-go/msgo/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMultipartMemory = 32 << 20

type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	engine     *Engine
	queryCache url.Values
	formCache  url.Values
}

func (c *Context) GetDefaultQuery(key, defaultValue string) string {
	values, ok := c.GetQueryArr(key)
	if !ok {
		return defaultValue
	}
	return values[0]
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) QueryArr(key string) []string {
	c.initQueryCache()
	values, _ := c.queryCache[key]
	return values
}

func (c *Context) GetQueryArr(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) QueryMap(key string) map[string]string {
	dict, _ := c.GetQueryMap(key)
	return dict
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) get(cache map[string][]string, key string) (map[string]string, bool) {
	// user[id]=1&user[name]=张三
	dict := make(map[string]string)
	exist := false
	for k, value := range cache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dict[k[i+1:][:j]] = value[0]
			}
		}
	}
	return dict, exist
}
func (c *Context) initFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

func (c *Context) GetPostFormArr(key string) ([]string, bool) {
	c.initFormCache()
	values, ok := c.formCache[key]
	return values, ok
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArr(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArr(key string) []string {
	values, _ := c.GetQueryArr(key)
	return values
}

func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

func (c *Context) FormFiles(name string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		log.Println(err)
	}
	return multipartForm.File[name]
}

func (c *Context) SaveUploadFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

func (c *Context) HTML(status int, html string) error {
	return c.Render(status, &render.HTML{
		Data:       html,
		IsTemplate: false,
	})
}

func (c *Context) HTMLTemplate(name string, data any, fileNames ...string) error {
	// 设置状态是200 默认不设置如果调用了 write 这个方法 实际上默认返回 200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseFiles(fileNames...)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	// 设置状态是200 默认不设置如果调用了 write 这个方法 实际上默认返回 200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) Template(name string, data any) error {
	return c.Render(http.StatusOK, &render.HTML{
		Name:       name,
		Data:       data,
		IsTemplate: true,
		Template:   c.engine.HTMLRender.Template,
	})
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(status, &render.JSON{
		Data: data,
	})
}

func (c *Context) XML(status int, data any) error {
	return c.Render(status, &render.XML{
		Data: data,
	})
}

func (c *Context) File(filePath string) {
	http.ServeFile(c.W, c.R, filePath)
}

func (c *Context) FileAttachment(filepath, filename string) {
	if IsASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// FileFromFS 参数中 filepath 是相对文件系统的路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)
	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(status int, url string) error {
	return c.Render(status, &render.Redirect{
		Code:     status,
		Request:  c.R,
		Location: url,
	})
}

func (c *Context) String(status int, format string, values ...any) error {
	return c.Render(status, &render.String{
		Format: format,
		Data:   values,
	})
}

func (c *Context) Render(statusCode int, r render.Render) error {
	err := r.Render(c.W)
	if statusCode != http.StatusOK {
		c.W.WriteHeader(statusCode)
	}
	return err
}
