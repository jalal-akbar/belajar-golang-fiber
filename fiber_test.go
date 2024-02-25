package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "embed"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

var app = fiber.New()

func TestRoutingHelloWorld(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(byte))
}

func TestRoutingCtx(t *testing.T) {
	app := fiber.New()
	app.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Query("name", "Guest")
		return c.SendString("Hello " + name)
	})
	req := httptest.NewRequest("GET", "/hello?name=Akbar", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Akbar", string(byte))
}

func TestHttpRequestFiber(t *testing.T) {
	app := fiber.New()
	app.Get("/request", func(c *fiber.Ctx) error {
		first := c.Get("firstname")
		last := c.Cookies("lastname")
		return c.SendString("Hello " + first + " " + last)
	})
	req := httptest.NewRequest("GET", "/request", nil)
	req.Header.Set("firstname", "Jalal")
	req.AddCookie(&http.Cookie{Name: "lastname", Value: "Akbar"})
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Jalal Akbar", string(byte))
}

func TestRouteParameterFiber(t *testing.T) {
	app := fiber.New()
	app.Get("/users/:userId/orders/:orderId", func(c *fiber.Ctx) error {
		userId := c.Params("userId")
		orderId := c.Params("orderId")
		return c.SendString("Get Order " + orderId + " from " + userId)
	})
	req := httptest.NewRequest("GET", "/users/Jalal/orders/2", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Get Order 2 from Jalal", string(byte))
}

func TestFormValueFiber(t *testing.T) {
	app := fiber.New()
	app.Post("/hello", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		return c.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Jalal")
	req := httptest.NewRequest("POST", "/hello", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Jalal", string(byte))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestMultipartFormFiber(t *testing.T) {
	app := fiber.New()
	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}

		err = c.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			return err
		}

		return c.SendString("Upload Success")
	})
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("file", "contoh.txt")
	assert.Nil(t, err)
	file.Write(contohFile)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload Success", string(byte))
}

// Request Body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {
	app := fiber.New()
	app.Post("/login", func(c *fiber.Ctx) error {
		body := c.Body()
		request := new(LoginRequest)

		err := json.Unmarshal(body, request)
		if err != nil {
			return err
		}
		return c.SendString("Hello " + request.Username)
	})
	body := strings.NewReader(`{"username":"akbar", "password":"rahasia"}`)
	request := httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	byte, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello akbar", string(byte))
}

// Body Parser
type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
	Name     string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T) {
	app.Post("/register", func(c *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := c.BodyParser(request)
		if err != nil {
			return err
		}
		return c.SendString("Register " + request.Username + " Success")
	})
}

func TestBodyParserJSON(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"akbar","password":"rahasia","name":"jalal"}`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(request)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register akbar Success", string(byte))
}

func TestBodyParserForm(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`username=akbar&password=rahasia&name=jalal`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := app.Test(request)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register akbar Success", string(byte))
}

func TestBodyParserXml(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(
		`<RegisterRequest>
			<username>akbar</username>
			<password>rahasia</password>
			<name>jalal</name>
		</RegisterRequest>`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/xml")
	resp, err := app.Test(request)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	byte, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register akbar Success", string(byte))
}

// HTTP Response

func TestJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/user", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{
			"username": "jalal",
			"name":     "jalal akbar",
		})
	})
	request := httptest.NewRequest("GET", "/user", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)

	body, err := io.ReadAll(response.Body)

	assert.Nil(t, err)
	assert.Equal(t, `{"name":"jalal akbar","username":"jalal"}`, string(body))
}

// Download File

func TestDownloadFile(t *testing.T) {

	t.Run("Download", func(t *testing.T) {
		app.Get("/download", func(c *fiber.Ctx) error {
			return c.Download("./source/contoh.txt", "contoh.txt")
		})
		request := httptest.NewRequest("GET", "/download", nil)
		response, err := app.Test(request)

		assert.Nil(t, err)
		assert.Equal(t, `attachment; filename="contoh.txt"`, response.Header.Get("Content-Disposition"))

		body, err := io.ReadAll(response.Body)

		assert.Nil(t, err)
		assert.Equal(t, "this is sample", string(body))
	})
	t.Run("Send", func(t *testing.T) {
		app.Get("/send", func(c *fiber.Ctx) error {
			return c.Send([]byte("Hello"))
		})
		request := httptest.NewRequest("GET", "/send", nil)
		response, err := app.Test(request)

		assert.Nil(t, err)

		body, err := io.ReadAll(response.Body)

		assert.Nil(t, err)
		assert.Equal(t, "Hello", string(body))
		fmt.Println("Send: ", string(body))
	})
	t.Run("SendFile", func(t *testing.T) {
		app.Get("/sendfile", func(c *fiber.Ctx) error {
			return c.SendFile("./source/contoh.txt")
		})
		request := httptest.NewRequest("GET", "/sendfile", nil)
		response, err := app.Test(request)

		assert.Nil(t, err)

		body, err := io.ReadAll(response.Body)

		assert.Nil(t, err)
		assert.Equal(t, "this is sample", string(body))
		fmt.Println("Send File: ", string(body))
	})

}

func TestRoutingGroup(t *testing.T) {
	helloWorld := func(c *fiber.Ctx) error {
		return c.SendString("Hello World")

	}

	api := app.Group("/api")
	api.Group("/hello", helloWorld)
	api.Group("/world", helloWorld)

	web := app.Group("/web")
	web.Group("/hello", helloWorld)
	web.Group("/world", helloWorld)

	request := httptest.NewRequest("GET", "/api/hello", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)

	body, err := io.ReadAll(response.Body)

	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(body))
}

//func (t *testing.T){}
