// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"os"
	path "path/filepath"
	"strings"
)

var cmdApiapp = &Command{
	// CustomFlags: true,
	UsageLine: "api [appname]",
	Short:     "create an API beego application",
	Long: `
Create an API beego application.

bee api [appname] [-tables=""] [-driver=mysql] [-conn=root:@tcp(127.0.0.1:3306)/test]
    -tables: a list of table names separated by ',' (default is empty, indicating all tables)
    -driver: [mysql | postgres | sqlite] (default: mysql)
    -conn:   the connection string used by the driver, the default is ''
             e.g. for mysql:    root:@tcp(127.0.0.1:3306)/test
             e.g. for postgres: postgres://postgres:postgres@127.0.0.1:5432/postgres

If 'conn' argument is empty, bee api creates an example API application,
when 'conn' argument is provided, bee api generates an API application based
on the existing database.

The command 'api' creates a folder named [appname] and inside the folder deploy
the following files/directories structure:

	├── conf
	│   └── app.conf
	├── controllers
	│   └── object.go
	│   └── user.go
	├── helpers
	│   └── global_function.go
	│   └── response_formater.go
	├── routers
	│   └── router.go
	├── tests
	│   └── default_test.go
	├── main.go
	└── models
	    └── object.go
	    └── user.go

`,
}

var apiconf = `appname = {{.Appname}}
httpaddr = "127.0.0.1"
httpport = 8080
runmode = "dev"
autorender = false
copyrequestbody = true
EnableDocs = true
mysqlurls = "127.0.0.1"
mysqluser = "dbuser"
mysqlpass = "dbpassword"
mysqldb   = "dbname"
`
var apiMaingo = `package main

import (
	_ "{{.Appname}}/docs"
	_ "{{.Appname}}/routers"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/plugins/cors"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	mysqlServer := beego.AppConfig.String("mysqlurls")
	mysqlUser := beego.AppConfig.String("mysqluser")
	mysqlPass := beego.AppConfig.String("mysqlpass")
	mysqlDb := beego.AppConfig.String("mysqldb")

	orm.RegisterDataBase("default", "mysql", mysqlUser+":"+mysqlPass+"@tcp("+mysqlServer+":3306)/"+mysqlDb)
	orm.DefaultTimeLoc = time.UTC

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "DELETE", "PUT", "PATCH", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}

func main() {
	if beego.RunMode == "dev" {
		beego.DirectoryIndex = true
		beego.StaticDir["/swagger"] = "swagger"
		orm.Debug = true
	}
	
	if beego.RunMode == "debug" {
		orm.Debug = true
	}

	orm.DefaultRelsDepth = 3

	beego.Run()
}
`

var apiMainconngo = `package main

import (
	_ "{{.Appname}}/docs"
	_ "{{.Appname}}/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	{{.DriverPkg}}
)

func init() {
	orm.RegisterDataBase("default", "{{.DriverName}}", "{{.conn}}")
}

func main() {
	if beego.RunMode == "dev" {
		beego.DirectoryIndex = true
		beego.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}

`

var apirouter = `// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"{{.Appname}}/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/object",
			beego.NSInclude(
				&controllers.ObjectController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
`

var apiModels = `package models

import (
	"errors"
	"strconv"
	"time"
)

var (
	Objects map[string]*Object
)

type Object struct {
	ObjectId   string
	Score      int64
	PlayerName string
}

func init() {
	Objects = make(map[string]*Object)
	Objects["hjkhsbnmn123"] = &Object{"hjkhsbnmn123", 100, "astaxie"}
	Objects["mjjkxsxsaa23"] = &Object{"mjjkxsxsaa23", 101, "someone"}
}

func AddOne(object Object) (ObjectId string) {
	object.ObjectId = "astaxie" + strconv.FormatInt(time.Now().UnixNano(), 10)
	Objects[object.ObjectId] = &object
	return object.ObjectId
}

func GetOne(ObjectId string) (object *Object, err error) {
	if v, ok := Objects[ObjectId]; ok {
		return v, nil
	}
	return nil, errors.New("ObjectId Not Exist")
}

func GetAll() map[string]*Object {
	return Objects
}

func Update(ObjectId string, Score int64) (err error) {
	if v, ok := Objects[ObjectId]; ok {
		v.Score = Score
		return nil
	}
	return errors.New("ObjectId Not Exist")
}

func Delete(ObjectId string) {
	delete(Objects, ObjectId)
}

`

var apiModels2 = `package models

import (
	"errors"
	"strconv"
	"time"
)

var (
	UserList map[string]*User
)

func init() {
	UserList = make(map[string]*User)
	u := User{"user_11111", "astaxie", "11111", Profile{"male", 20, "Singapore", "astaxie@gmail.com"}}
	UserList["user_11111"] = &u
}

type User struct {
	Id       string
	Username string
	Password string
	Profile  Profile
}

type Profile struct {
	Gender  string
	Age     int
	Address string
	Email   string
}

func AddUser(u User) string {
	u.Id = "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	UserList[u.Id] = &u
	return u.Id
}

func GetUser(uid string) (u *User, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}

func GetAllUsers() map[string]*User {
	return UserList
}

func UpdateUser(uid string, uu *User) (a *User, err error) {
	if u, ok := UserList[uid]; ok {
		if uu.Username != "" {
			u.Username = uu.Username
		}
		if uu.Password != "" {
			u.Password = uu.Password
		}
		if uu.Profile.Age != 0 {
			u.Profile.Age = uu.Profile.Age
		}
		if uu.Profile.Address != "" {
			u.Profile.Address = uu.Profile.Address
		}
		if uu.Profile.Gender != "" {
			u.Profile.Gender = uu.Profile.Gender
		}
		if uu.Profile.Email != "" {
			u.Profile.Email = uu.Profile.Email
		}
		return u, nil
	}
	return nil, errors.New("User Not Exist")
}

func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}

func DeleteUser(uid string) {
	delete(UserList, uid)
}
`

var apiControllers = `package controllers

import (
	"{{.Appname}}/models"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about object
type ObjectController struct {
	beego.Controller
}

// @Title create
// @Description create object
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} models.Object.Id
// @Failure 403 body is empty
// @router / [post]
func (o *ObjectController) Post() {
	var ob models.Object
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)
	objectid := models.AddOne(ob)
	o.Data["json"] = map[string]string{"ObjectId": objectid}
	o.ServeJson()
}

// @Title Get
// @Description find object by objectid
// @Param	objectId		path 	string	true		"the objectid you want to get"
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router /:objectId [get]
func (o *ObjectController) Get() {
	objectId := o.Ctx.Input.Params[":objectId"]
	if objectId != "" {
		ob, err := models.GetOne(objectId)
		if err != nil {
			o.Data["json"] = err
		} else {
			o.Data["json"] = ob
		}
	}
	o.ServeJson()
}

// @Title GetAll
// @Description get all objects
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router / [get]
func (o *ObjectController) GetAll() {
	obs := models.GetAll()
	o.Data["json"] = obs
	o.ServeJson()
}

// @Title update
// @Description update the object
// @Param	objectId		path 	string	true		"The objectid you want to update"
// @Param	body		body 	models.Object	true		"The body"
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router /:objectId [put]
func (o *ObjectController) Put() {
	objectId := o.Ctx.Input.Params[":objectId"]
	var ob models.Object
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)

	err := models.Update(objectId, ob.Score)
	if err != nil {
		o.Data["json"] = err
	} else {
		o.Data["json"] = "update success!"
	}
	o.ServeJson()
}

// @Title delete
// @Description delete the object
// @Param	objectId		path 	string	true		"The objectId you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 objectId is empty
// @router /:objectId [delete]
func (o *ObjectController) Delete() {
	objectId := o.Ctx.Input.Params[":objectId"]
	models.Delete(objectId)
	o.Data["json"] = "delete success!"
	o.ServeJson()
}

`
var apiControllers2 = `package controllers

import (
	"{{.Appname}}/models"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Users
type UserController struct {
	beego.Controller
}

// @Title createUser
// @Description create users
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} models.User.Id
// @Failure 403 body is empty
// @router / [post]
func (u *UserController) Post() {
	var user models.User
	json.Unmarshal(u.Ctx.Input.RequestBody, &user)
	uid := models.AddUser(user)
	u.Data["json"] = map[string]string{"uid": uid}
	u.ServeJson()
}

// @Title Get
// @Description get all Users
// @Success 200 {object} models.User
// @router / [get]
func (u *UserController) GetAll() {
	users := models.GetAllUsers()
	u.Data["json"] = users
	u.ServeJson()
}

// @Title Get
// @Description get user by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.User
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *UserController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		user, err := models.GetUser(uid)
		if err != nil {
			u.Data["json"] = err
		} else {
			u.Data["json"] = user
		}
	}
	u.ServeJson()
}

// @Title update
// @Description update the user
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {object} models.User
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *UserController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var user models.User
		json.Unmarshal(u.Ctx.Input.RequestBody, &user)
		uu, err := models.UpdateUser(uid, &user)
		if err != nil {
			u.Data["json"] = err
		} else {
			u.Data["json"] = uu
		}
	}
	u.ServeJson()
}

// @Title delete
// @Description delete the user
// @Param	uid		path 	string	true		"The uid you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (u *UserController) Delete() {
	uid := u.GetString(":uid")
	models.DeleteUser(uid)
	u.Data["json"] = "delete success!"
	u.ServeJson()
}

// @Title login
// @Description Logs user into the system
// @Param	username		query 	string	true		"The username for login"
// @Param	password		query 	string	true		"The password for login"
// @Success 200 {string} login success
// @Failure 403 user not exist
// @router /login [get]
func (u *UserController) Login() {
	username := u.GetString("username")
	password := u.GetString("password")
	if models.Login(username, password) {
		u.Data["json"] = "login success"
	} else {
		u.Data["json"] = "user not exist"
	}
	u.ServeJson()
}

// @Title logout
// @Description Logs out current logged in user session
// @Success 200 {string} logout success
// @router /logout [get]
func (u *UserController) Logout() {
	u.Data["json"] = "logout success"
	u.ServeJson()
}

`

var apiTests = `package test


import (
	"net/http"
	"net/http/httptest"
	"testing"
	"runtime"
	"path/filepath"
	_ "{{.Appname}}/routers"

	"github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

// TestGet is a sample to run an endpoint test
func TestGet(t *testing.T) {
	r, _ := http.NewRequest("GET", "/v1/object", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Test Station Endpoint\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})
}

`

var apiGlobalFunction = `package helpers

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
)

// Get input keys
func GetInputKeys(input []byte) []string {
	// convert input into map
	var objmap map[string]*json.RawMessage // ambil input jadikan map
	json.Unmarshal(input, &objmap)

	// get the keys
	// ambil keys ganti jadi slice
	keys := make([]string, 0, len(objmap))
	for k := range objmap {
		keys = append(keys, k)
	}
	return keys
}

// function validator
func Validator(model interface{}) (bool, map[string]interface{}) {

	errorData := make(map[string]string)
	valid := validation.Validation{}

	passed, _ := valid.Valid(model)
	if !passed {
		for _, err := range valid.Errors {
			field := strings.Split(err.Key, ".")
			errorData[field[0]] = err.Message
		}
		Rf.Fail(errorData)
		return false, Rf.Data
	} else {
		// disini ntar format data unutk sukses
		return true, Rf.Data
	}
}

func QueryString(qs url.Values) (query map[int]map[string]string, fields []string, groupby []string, sortby []string, order []string,
	offset int64, limit int64) {

	var cq map[string]string = make(map[string]string)
	query = make(map[int]map[string]string)
	limit = 10
	offset = 0

	for k, v := range qs {

		// if value or maps not empty, added the value to variable
		if v[0] != "" {
			if k == "fields" {
				fields = strings.Split(v[0], ",")
			} else if k == "groupby" {
				groupby = strings.Split(v[0], ",")
			} else if k == "sortby" {
				sortby = strings.Split(v[0], ",")
			} else if k == "order" {
				order = strings.Split(v[0], ",")
			} else if k == "limit" {
				x, _ := strconv.Atoi(v[0])
				limit = int64(x)
			} else if k == "offset" {
				x, _ := strconv.Atoi(v[0])
				offset = int64(x)
			} else if k == "query" {
				var index int = 0

				// query string with multi condition
				// field1:x|field2:y,Or.field2:z
				// field1 = x and (field2=y or field2=z)
				for _, cond := range strings.Split(v[0], "|") {

					for _, partcond := range strings.Split(cond, ",") {
						kv := strings.Split(partcond, ":")
						if len(kv) != 2 {
							return query, fields, groupby, sortby, order, offset, limit
						}
						k, val := kv[0], kv[1]
						cq[k] = val
					}

					index = index + 1
					query[index] = cq

					// reset the map
					cq = make(map[string]string)
				}
			}
		}
	}

	return query, fields, groupby, sortby, order, offset, limit
}

func QueryCondition(query map[int]map[string]string) (cond *orm.Condition) {
	cond = orm.NewCondition()
	condition := orm.NewCondition()

	for _, q := range query {
		condition = orm.NewCondition()

		for k, v := range q {
			if strings.Contains(k, "And.") {
				k = strings.Replace(k, "And.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = condition.And(k, v)
			} else if strings.Contains(k, "IsNull.") {
				k = strings.Replace(k, "IsNull.", "", -1)
				k = strings.Replace(k, ".", "__", -1)
				k += "__isnull"

				condition = condition.And(k, true)
			} else if strings.Contains(k, "NotNull.") {
				k = strings.Replace(k, "NotNull.", "", -1)
				k = strings.Replace(k, ".", "__", -1)
				k += "__isnull"

				condition = condition.And(k, false)
			} else if strings.Contains(k, "Ex.") {
				k = strings.Replace(k, "Ex.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = condition.AndNot(k, v)
			} else if strings.Contains(k, "Or.") {
				k = strings.Replace(k, "Or.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = condition.Or(k, v)
			} else if strings.Contains(k, "OrNot.") {
				k = strings.Replace(k, "OrNot.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = condition.OrNot(k, v)
			} else if strings.Contains(k, ".in") {
				k = strings.Replace(k, ".", "__", -1)
				vArr := strings.Split(v, ".")

				condition = condition.And(k, vArr)
			} else {
				k = strings.Replace(k, ".", "__", -1)

				condition = condition.And(k, v)
			}
		}

		// merge with AND
		// @todo need make like this for OR
		cond = cond.AndCond(condition)
	}

	return cond
}
`

var apiResponseFormater = `package helpers

import (
	"strings"
)

type ResponseFormat struct {
	Data map[string]interface{}
}

// global response formatter
var (
	Rf = ResponseFormat{}
)

func (r *ResponseFormat) Success(httpMethod string, id int, d ...[]interface{}) {
	switch httpMethod {
	case "POST":
		r.Data = make(map[string]interface{})
		r.Data["success"] = true
		r.Data["id"] = id
	case "GET":
		if d != nil {
			r.Data["data"] = d[0]
		}
	default:
		r.Data = make(map[string]interface{})
		r.Data["success"] = true
	}

}

func (r *ResponseFormat) Fail(errorData interface{}) {
	r.Data = make(map[string]interface{})
	r.Data["success"] = false

	switch errorData.(type) {
	case string:
		r.Data["error"] = map[string]string{
			"orm": ClearErrorPrefix(errorData.(string)), // from orm
		}
	default:
		r.Data["error"] = errorData // from validator
	}
}

// Beego "No row" SQL error has "<QuerySetter>" prefix which is really annoying
// remove it with this function
func ClearErrorPrefix(s string) string {
	strToRemove := "<QuerySeter> "
	s = strings.TrimPrefix(s, strToRemove)
	return s
}
`

func init() {
	cmdApiapp.Run = createapi
	cmdApiapp.Flag.Var(&tables, "tables", "specify tables to generate model")
	cmdApiapp.Flag.Var(&driver, "driver", "database driver: mysql, postgresql, etc.")
	cmdApiapp.Flag.Var(&conn, "conn", "connection string used by the driver to connect to a database instance")
}

func createapi(cmd *Command, args []string) int {
	curpath, _ := os.Getwd()
	if len(args) < 1 {
		ColorLog("[ERRO] Argument [appname] is missing\n")
		os.Exit(2)
	}
	if len(args) > 1 {
		cmd.Flag.Parse(args[1:])
	}
	apppath, packpath, err := checkEnv(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if driver == "" {
		driver = "mysql"
	}
	if conn == "" {
	}
	os.MkdirAll(apppath, 0755)
	fmt.Println("create app folder:", apppath)
	os.Mkdir(path.Join(apppath, "conf"), 0755)
	fmt.Println("create conf:", path.Join(apppath, "conf"))
	os.Mkdir(path.Join(apppath, "controllers"), 0755)
	fmt.Println("create controllers:", path.Join(apppath, "controllers"))
	os.Mkdir(path.Join(apppath, "docs"), 0755)
	fmt.Println("create docs:", path.Join(apppath, "docs"))
	os.Mkdir(path.Join(apppath, "tests"), 0755)
	fmt.Println("create helpers:", path.Join(apppath, "helpers"))
	os.Mkdir(path.Join(apppath, "helpers"), 0755)
	fmt.Println("create tests:", path.Join(apppath, "tests"))

	fmt.Println("create conf app.conf:", path.Join(apppath, "conf", "app.conf"))
	writetofile(path.Join(apppath, "conf", "app.conf"),
		strings.Replace(apiconf, "{{.Appname}}", args[0], -1))

	fmt.Println("create file global_function.go:", path.Join(apppath, "helpers", "global_function.go"))
	writetofile(path.Join(apppath, "helpers", "global_function.go"),
		strings.Replace(apiGlobalFunction, "{{.Appname}}", args[0], -1))

	fmt.Println("create file response_formater.go:", path.Join(apppath, "helpers", "response_formater.go"))
	writetofile(path.Join(apppath, "helpers", "response_formater.go"),
		strings.Replace(apiResponseFormater, "{{.Appname}}", args[0], -1))

	if conn != "" {
		fmt.Println("create main.go:", path.Join(apppath, "main.go"))
		maingoContent := strings.Replace(apiMainconngo, "{{.Appname}}", packpath, -1)
		maingoContent = strings.Replace(maingoContent, "{{.DriverName}}", string(driver), -1)
		if driver == "mysql" {
			maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/go-sql-driver/mysql"`, -1)
		} else if driver == "postgres" {
			maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/lib/pq"`, -1)
		}
		writetofile(path.Join(apppath, "main.go"),
			strings.Replace(
				maingoContent,
				"{{.conn}}",
				conn.String(),
				-1,
			),
		)
		ColorLog("[INFO] Using '%s' as 'driver'\n", driver)
		ColorLog("[INFO] Using '%s' as 'conn'\n", conn)
		ColorLog("[INFO] Using '%s' as 'tables'\n", tables)
		generateAppcode(string(driver), string(conn), "3", string(tables), path.Join(curpath, args[0]))
	} else {
		os.Mkdir(path.Join(apppath, "models"), 0755)
		fmt.Println("create models:", path.Join(apppath, "models"))
		os.Mkdir(path.Join(apppath, "routers"), 0755)
		fmt.Println(path.Join(apppath, "routers") + string(path.Separator))

		fmt.Println("create controllers object.go:", path.Join(apppath, "controllers", "object.go"))
		writetofile(path.Join(apppath, "controllers", "object.go"),
			strings.Replace(apiControllers, "{{.Appname}}", packpath, -1))

		fmt.Println("create controllers user.go:", path.Join(apppath, "controllers", "user.go"))
		writetofile(path.Join(apppath, "controllers", "user.go"),
			strings.Replace(apiControllers2, "{{.Appname}}", packpath, -1))

		fmt.Println("create tests default.go:", path.Join(apppath, "tests", "default_test.go"))
		writetofile(path.Join(apppath, "tests", "default_test.go"),
			strings.Replace(apiTests, "{{.Appname}}", packpath, -1))

		fmt.Println("create routers router.go:", path.Join(apppath, "routers", "router.go"))
		writetofile(path.Join(apppath, "routers", "router.go"),
			strings.Replace(apirouter, "{{.Appname}}", packpath, -1))

		fmt.Println("create models object.go:", path.Join(apppath, "models", "object.go"))
		writetofile(path.Join(apppath, "models", "object.go"), apiModels)

		fmt.Println("create models user.go:", path.Join(apppath, "models", "user.go"))
		writetofile(path.Join(apppath, "models", "user.go"), apiModels2)

		fmt.Println("create docs doc.go:", path.Join(apppath, "docs", "doc.go"))
		writetofile(path.Join(apppath, "docs", "doc.go"), "package docs")

		fmt.Println("create main.go:", path.Join(apppath, "main.go"))
		writetofile(path.Join(apppath, "main.go"),
			strings.Replace(apiMaingo, "{{.Appname}}", packpath, -1))
	}
	return 0
}

func checkEnv(appname string) (apppath, packpath string, err error) {
	curpath, err := os.Getwd()
	if err != nil {
		return
	}

	gopath := os.Getenv("GOPATH")
	Debugf("gopath:%s", gopath)
	if gopath == "" {
		err = fmt.Errorf("you should set GOPATH in the env")
		return
	}

	appsrcpath := ""
	haspath := false
	wgopath := path.SplitList(gopath)
	for _, wg := range wgopath {
		wg, _ = path.EvalSymlinks(path.Join(wg, "src"))

		if path.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}
	}

	if !haspath {
		err = fmt.Errorf("can't create application outside of GOPATH `%s`\n"+
			"you first should `cd $GOPATH%ssrc` then use create\n", gopath, string(path.Separator))
		return
	}
	apppath = path.Join(curpath, appname)

	if _, e := os.Stat(apppath); os.IsNotExist(e) == false {
		err = fmt.Errorf("path `%s` exists, can not create app without remove it\n", apppath)
		return
	}
	packpath = strings.Join(strings.Split(apppath[len(appsrcpath)+1:], string(path.Separator)), "/")
	return
}
