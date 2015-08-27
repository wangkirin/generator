package middleware

import (
	"html/template"

	"github.com/Unknwon/macaron"
	_ "github.com/macaron-contrib/session/redis"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	m.Use(macaron.Recovery())

	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger())

	//modify  default template setting
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:       "views",
		Extensions:      []string{".tmpl", ".html"},
		Funcs:           []template.FuncMap{},
		Delims:          macaron.Delims{"<<<", ">>>"},
		Charset:         "UTF-8",
		IndentJSON:      true,
		IndentXML:       true,
		PrefixXML:       []byte("macaron"),
		HTMLContentType: "text/html",
	}))
}
