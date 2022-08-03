package main

import (
	"fmt"
	"github.com/demo-go/msgo"
	"log"
	"net/http"
)

type User struct {
	Name string
	Age  int
}

func Log(next msgo.HandleFunc) msgo.HandleFunc {
	return func(ctx *msgo.Context) {
		fmt.Println("666666")
		next(ctx)
		fmt.Println("999999")
	}
}

func main() {
	engine := msgo.New()
	g := engine.Group("user")
	g.Use(func(next msgo.HandleFunc) msgo.HandleFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("use pre middleware")
			next(ctx)
			fmt.Println("use post middleware")
		}
	})
	g.Get("/hello", func(ctx *msgo.Context) {
		fmt.Println("handle")
		fmt.Fprintf(ctx.W, "%s get 欢迎来到码神之路goweb教程", "dema-go.com")
	}, Log)
	g.Get("/hello/get", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s /hello/*/get 欢迎来到码神之路goweb教程", "dema-go.com")
	})
	g.Post("/hello", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s post 欢迎来到码神之路goweb教程", "dema-go.com")
	})
	g.Post("/info", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s info", "dema-go.com")
	})
	g.Any("/any", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s any", "dema-go.com")
	})
	g.Get("/get/:id", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s get user info path variable", "dema-go.com")
	})

	g.Get("/html", func(ctx *msgo.Context) {
		err := ctx.HTML(http.StatusOK, "<h1>666</h1>")
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/htmlTemplate", func(ctx *msgo.Context) {
		user := &User{
			Name: "dema",
			Age:  19,
		}
		err := ctx.HTMLTemplate("login.html", user, "./tpl/login.html", "./tpl/header.html")
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/htmlTemplateGlob", func(ctx *msgo.Context) {
		user := &User{
			Name: "dema",
			Age:  19,
		}
		err := ctx.HTMLTemplateGlob("login.html", user, "./tpl/*")
		if err != nil {
			log.Println(err)
		}
	})
	engine.LoadTemplate("./tpl/*")
	g.Get("/template", func(ctx *msgo.Context) {
		user := &User{
			Name: "dema",
		}
		err := ctx.Template("login.html", user)
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/json", func(ctx *msgo.Context) {
		user := &User{
			Name: "dema",
			Age:  19,
		}
		err := ctx.JSON(http.StatusOK, user)
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/xml", func(ctx *msgo.Context) {
		user := &User{
			Name: "dema",
		}
		err := ctx.XML(http.StatusOK, user)
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/excel", func(ctx *msgo.Context) {
		ctx.File("./tpl/test.xlsx")
	})
	g.Get("/excelName", func(ctx *msgo.Context) {
		ctx.FileAttachment("./tpl/test.xlsx", "dema666")
	})
	g.Get("/fs", func(ctx *msgo.Context) {
		ctx.FileFromFS("test.xlsx", http.Dir("./tpl"))
	})
	g.Get("/redirect", func(ctx *msgo.Context) {
		ctx.Redirect(http.StatusFound, "/user/template")
	})
	g.Get("/string", func(ctx *msgo.Context) {
		err := ctx.String(http.StatusFound, "%s 和 %s 666", "dema", "xiya")
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/add", func(ctx *msgo.Context) {
		name := ctx.GetDefaultQuery("name", "dema")
		fmt.Printf("name: %s\n", name)
	})
	g.Get("/queryMap", func(ctx *msgo.Context) {
		m, _ := ctx.GetQueryMap("user")
		ctx.JSON(http.StatusOK, m)
	})
	g.Post("/formPost", func(ctx *msgo.Context) {
		name, _ := ctx.GetPostFormMap("user")
		ctx.JSON(http.StatusOK, name)
	})
	engine.Run()
}
