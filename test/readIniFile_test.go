package test

import (
	"testing"

	"github.com/astaxie/beego/config"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	conf config.ConfigContainer
	err  error
)

func TestReadIniFile(t *testing.T) {
	Convey("Read a ini file", t, func() {
		conf, err = config.NewConfig("ini", "config.ini")
		So(err, ShouldBeNil)
		So(conf, ShouldNotBeNil)
	})
	Convey("Read a value responding to the key from ini file", t, func() {
		Convey("Read appname", func() {
			appname := conf.String("appname")
			So(appname, ShouldEqual, "generator")
		})
		Convey("Read usage", func() {
			usage := conf.String("usage")
			So(usage, ShouldEqual, "Generator Service")
		})
		Convey("Read version", func() {
			version := conf.String("version")
			So(version, ShouldEqual, "0.0.1")
		})
		Convey("Read author", func() {
			author := conf.String("author")
			So(author, ShouldEqual, "Meaglith Ma")
		})
		Convey("Read email", func() {
			email := conf.String("email")
			So(email, ShouldEqual, "genedna@gmail.com")
		})
		Convey("Read db::uri", func() {
			dburi := conf.String("db::uri")
			So(dburi, ShouldEqual, "localhost:6379")
		})
		Convey("Read db::passwd", func() {
			dbpasswd := conf.String("db::passwd")
			So(dbpasswd, ShouldEqual, "containerops")
		})
		Convey("Read db:db", func() {
			dbdb, err := conf.Int64("db::db")
			So(dbdb, ShouldEqual, 2)
			So(err, ShouldBeNil)
		})
	})
}
