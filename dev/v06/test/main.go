package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cainmusic/rough/dev/v06"
)

// 为保证可以读取到templates，需要进入"dev/v06/test"目录下执行

func main() {
	r := rough.New()

	// test v06 start

	// 设置html template前设置路由
	r.GET("/before_html_template", quickUrl)

	// 加载全局模板
	// 注：使用"template/*"会无法解析"template/files"目录，所以通配符用"*.html"
	//     多级目录用"**/**/*"来匹配，但仅能匹配到"*"级目录中的文件
	//     多级匹配的文件名前面要带匹配的路径（或者在模板内重定义模板名）
	r.LoadHTMLGlob("../template/*.html")

	r.GET("/visit_index", func(c *rough.Context) {
		c.HTML(http.StatusOK, "index.html", rough.H{"title": "你好"})
	})

	/*
		// LoadHTMLGlob和LoadHTMLFiles仅一次生效
		// 加载文件模板
		r.LoadHTMLFiles("../template/files/1.html", "template/files/2.html")

		r.GET("/visit_files_1", func(c *rough.Context) {
			c.HTML(http.StatusOK, "1.html", rough.H{"title": "你好，这是1"})
		})

		r.GET("/visit_files_2", func(c *rough.Context) {
			c.HTML(http.StatusOK, "2.html", rough.H{"title": "你好，这是2"})
		})
	*/

	r.GET("/some_json", func(c *rough.Context) {
		c.JSON(http.StatusOK, rough.H{"name": "小王", "age": 18})
	})

	r.GET("/some_json_2", func(c *rough.Context) {
		c.JSON(http.StatusOK, struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{
			"小张",
			17,
		})
	})

	r.GET("/redirect", func(c *rough.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/visit_index")
	})

	r.GET("/test_hhh", func(c *rough.Context) {
		c.String(http.StatusOK, "abc")
		c.String(http.StatusOK, "def")
	})

	// test v06 end

	middlewares(r)

	routes(r)

	routergroupG1(r)

	r.Run()
}

func middlewares(r *rough.Engine) {
	r.Use(func(c *rough.Context) {
		st := time.Now()
		c.Set("tnow", st)
		log.Println("start time", st.UnixNano())
		c.Next()
		et := time.Now()
		log.Println("end time", et.UnixNano())
		log.Println("use time", et.UnixMicro()-st.UnixMicro(), "microsecond")
	})

	r.Use(func(c *rough.Context) {
		log.Println("middleware without calling next")
	})

	r.Use(func(c *rough.Context) {
		t, err := c.Get("tnow")
		if err != nil {
			panic("ERROR TODO : " + err.Error())
		}
		//c.String(
		//	http.StatusOK,
		log.Printf(
			"hello, rough serves you. cur url is %s, method is %s, request time is %s",
			c.R.URL.Path,
			c.R.Method,
			t.(time.Time).String(),
		)
	})
}

func routes(r *rough.Engine) {
	r.GET("/", func(c *rough.Context) {
		c.String(http.StatusOK, "use router")
	})

	r.Any("/any", func(c *rough.Context) {
		c.String(http.StatusOK, "you can use any method to request this url")
	})

	r.GET("/query", func(c *rough.Context) {
		log.Println(c.Query("a"))
		log.Println(c.Query("b"))
		bSlice, _ := c.GetQueryArray("b")
		log.Println(bSlice)
	}, quickUrl)

	r.POST("/post", func(c *rough.Context) {
		log.Println(c.Query("a"))
		log.Println(c.Query("b"))
		bSlice, _ := c.GetQueryArray("b")
		log.Println(bSlice)

		log.Println("request form before parse", c.R.Form["a"])
		log.Println("request form before parse", c.R.Form["b"])

		log.Println("request post form not contain query a", c.PostForm("a"))
		log.Println("request post form not contain query b", c.PostForm("b"))
		log.Println("request post form c", c.PostForm("c"))
		log.Println("request post form d", c.PostForm("d"))
		dSlice, _ := c.GetPostFormArray("d")
		log.Println("request post form d array", dSlice)

		log.Println("request after before parse", c.R.Form["a"])
		log.Println("request after before parse", c.R.Form["b"])
	}, quickUrl)
}

func routergroupG1(r *rough.Engine) {
	g1 := r.Group("/g1", func(c *rough.Context) {
		log.Println("this is a first middleware for group g1 with next called")
		c.Next()
		log.Println("end of the first middleware")
	}, func(c *rough.Context) {
		log.Println("this is a second middleware for group g1")
	}, func(c *rough.Context) {
		log.Println("this is a third middleware for group g1")
	})
	g1.GET("/abc", func(c *rough.Context) {
		c.String(http.StatusOK, "response from handler for /g1/abc")
	})
}

func quickUrl(c *rough.Context) {
	c.String(http.StatusOK, "you just visited "+c.R.URL.String())
}
