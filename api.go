package main

import (
	"Cloud/store"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type api struct {
	cfg   config
	store *store.Storage
}

type config struct {
	addr              string
	echoTimeout       time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	requestLimitation int
	jwtSecret         []byte
	adminWL           []string
}

func HTTPErrorHandler(err error, c echo.Context) {
	var code int
	var message string
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	if code == http.StatusInternalServerError {
		c.File("web/500.html")
		return
	} else if code == http.StatusNotFound {
		c.File("web/404.html")
		return
	} else if code == http.StatusForbidden {
		c.File("web/403.html")
		return
	}
	c.JSON(code, map[string]string{
		"error": message,
	})
}

func (api *api) Authentificator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.Contains(c.Path(), "/style/") || c.Path() == "/logout" {
			return next(c)
		}
		err := api.AuthValidate(&c)
		if err != nil {
			if c.Path() == "/login" {
				return next(c)
			}
			return c.Redirect(http.StatusFound, "/login")
		}

		if c.Path() == "/login" || c.Path() == "/" {
			return c.Redirect(http.StatusFound, "/dashboard")
		}
		if c.Path() == "/admin" || strings.Contains(c.Path(), "/users") {
			user := c.Get("user").(*store.User)
			if user.Permissions != 1 || !slices.Contains(api.cfg.adminWL, c.RealIP()) {
				return c.File("web/403.html")
			}
		}

		return next(c)
	}
}

func (api *api) Mount() *echo.Echo {
	e := echo.New()
	e.Use(api.Authentificator)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: api.cfg.echoTimeout,
	}))
	e.Use(middleware.BodyLimit("20G"))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(api.cfg.requestLimitation))))
	e.HTTPErrorHandler = HTTPErrorHandler
	server := &http.Server{
		Addr:           api.cfg.addr,
		Handler:        e,
		ReadTimeout:    api.cfg.readTimeout,
		WriteTimeout:   api.cfg.writeTimeout,
		IdleTimeout:    api.cfg.idleTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	e.Server = server

	e.GET("/", api.LoginHandler)

	e.GET("/login", api.LoginHandler)
	e.POST("/login", api.LoginHandler)

	e.GET("/dashboard", api.DashboardHandler)
	e.GET("/admin", api.AdminHandler)

	e.GET("/style/:filepath", api.StyleHandler)
	e.GET("/profile", api.ProfileHandler)

	e.GET("/users", api.UsersHandler)
	e.POST("/users", api.UsersHandler)
	e.PUT("/users", api.UsersHandler)
	e.DELETE("/users", api.UsersHandler)

	e.GET("/files", api.FilesHandler)
	e.POST("/files", api.FilesHandler)
	e.PUT("/files", api.FilesHandler)
	e.DELETE("/files", api.FilesHandler)

	e.GET("/logout", api.LogoutHandler)

	return e
}

func (api *api) Run() error {
	e := api.Mount()
	return e.Start(api.cfg.addr)
}

func (api *api) LoginHandler(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		return c.File("web/login.html")
	} else if c.Request().Method == http.MethodPost {
		login := strings.Trim(c.FormValue("login"), " ")
		password := c.FormValue("password")
		user, err := api.store.Users.GetByName(login)

		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}
		err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":  user.Id,
			"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
		})

		tokenString, err := token.SignedString(api.cfg.jwtSecret)
		if err != nil {
			return c.String(http.StatusInternalServerError, "ERR_TOKEN_SIGN")
		}
		c.SetCookie(&http.Cookie{
			Name:     "Auth",
			Value:    tokenString,
			MaxAge:   3600 * 24 * 30,
			HttpOnly: true,
		})

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.String(http.StatusBadRequest, "ERR_BAD_METHOD")
}

func (api *api) DashboardHandler(c echo.Context) error {
	return c.File("web/dashboard.html")
}

func (api *api) AdminHandler(c echo.Context) error {
	return c.File("web/admin.html")
}

func (api *api) UsersHandler(c echo.Context) error {
	params, err := c.FormParams()
	if err != nil || params == nil {
		return c.String(http.StatusBadRequest, "ERR_BAD")
	}

	if c.Request().Method == http.MethodGet {
		p, err := strconv.Atoi(params.Get("page"))
		s, err := strconv.Atoi(params.Get("size"))

		u, err := api.store.Users.GetPage(p, s)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}
		return c.JSON(http.StatusOK, u)
	}

	if c.Request().Method == http.MethodPost {
		name := params.Get("name")
		password, err := bcrypt.GenerateFromPassword([]byte(params.Get("password")), bcrypt.DefaultCost)
		space, err := strconv.Atoi(params.Get("space"))
		if err != nil || name == "" || params.Get("password") == "" {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}
		err = api.store.Users.Create(store.NewUser(name, password, 0, float32(space*1024)))

		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}
		return c.String(http.StatusOK, "OK")
	}

	if c.Request().Method == http.MethodDelete {
		id, err := strconv.Atoi(params.Get("id"))
		err = api.store.Users.Delete(id)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}

		os.Remove(fmt.Sprintf("uploads/profile_pictures/%d.jpg", id))
		return c.String(http.StatusOK, "OK")
	}

	if c.Request().Method == http.MethodPut {
		id, err := strconv.Atoi(params.Get("id"))
		u, err := api.store.Users.GetById(id)
		name := params.Get("name")
		password := params.Get("password")
		space := params.Get("space")
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}

		if name != "" {
			u.Username = name
		}
		if password != "" {
			passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			u.Password = passwordHash
		}
		if space != "" {
			spaceInt, _ := strconv.Atoi(space)
			u.Space = float32(spaceInt * 1024)
		}

		err = api.store.Users.Update(u)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_BAD")
		}
		return c.String(http.StatusOK, "OK")
	}

	return c.String(http.StatusBadRequest, "ERR_BAD")
}

func (api *api) StyleHandler(c echo.Context) error {
	path := "web/style/" + c.Param("filepath")
	abs, err := filepath.Abs(path)
	_, err = os.Stat(abs)

	if err != nil {
		if strings.Contains(path, "/file_icons/") {
			return c.File("web/style/icons/file_icons/file.png")
		}
		return c.String(http.StatusNotFound, "ERR_NOT_FOUND")
	}

	pathDirs := strings.Split(abs, "\\")
	if len(pathDirs) == 1 {
		pathDirs = strings.Split(strings.Join(pathDirs, ""), "/") // for linux systems where path uses / instead of \
	}
	if slices.Contains(pathDirs, "web") && slices.Contains(pathDirs, "style") {
		return c.File(path)
	} else {
		return c.String(http.StatusNotFound, "ERR_NOT_FOUND")
	}
}

func (api *api) ProfileHandler(c echo.Context) error {
	user := c.Get("user").(*store.User)

	param := c.QueryParam("get")
	if param == "" {
		spaceOccupied, _ := api.store.Files.GetOccupiedSpace(user.Id)
		res := map[string]interface{}{
			"id":              user.Id,
			"name":            user.Username,
			"space_available": user.Space,
			"space_occupied":  spaceOccupied,
		}
		return c.JSON(http.StatusOK, res)
	} else if param == "picture" {
		id := c.QueryParam("id")
		if id != "" {
			if user.Permissions != 1 {
				return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
			}

			path := fmt.Sprintf("uploads/profile_pictures/%s.jpg", id)
			if _, err := os.Stat(path); err != nil {
				return c.File("web/style/img/placeholder.jpg")
			}

			abs, _ := filepath.Abs(path)
			if strings.Contains(abs, "profile_pictures") {
				return c.File(path)
			} else {
				return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
			}
		}

		path := fmt.Sprintf("uploads/profile_pictures/%d.jpg", user.Id)
		if _, err := os.Stat(path); err != nil {
			return c.File("web/style/img/placeholder.jpg")
		}
		return c.File(path)
	}

	return c.String(http.StatusBadRequest, "ERR_BAD")
}

func (api *api) AuthValidate(c *echo.Context) error {
	authCookie, err := (*c).Cookie("Auth")
	if err != nil {
		(*c).SetCookie(&http.Cookie{
			Name:     "Auth",
			Value:    "",
			MaxAge:   3600 * 24 * 30,
			HttpOnly: true,
		})
		authCookie, err = (*c).Cookie("Auth")
		return (*c).Redirect(http.StatusOK, "/login")
	}

	parsedToken, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("ERR_AUTH_SIGN")
		}
		return api.cfg.jwtSecret, nil
	})
	if err != nil {
		return errors.New("ERR_AUTH_INVALID")
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return errors.New("ERR_AUTH_EXPIRED")
		}

		user, err := api.store.Users.GetById(int(claims["id"].(float64)))
		if err != nil {
			return err
		}
		(*c).Set("user", user)
		return nil
	} else {
		return errors.New("ERR_AUTH_INVALID")
	}
}

func (api *api) LogoutHandler(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "Auth",
		Value:    "",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
	})
	return c.Redirect(http.StatusFound, "/login")
}

func (api *api) FilesHandler(c echo.Context) error {
	user := c.Get("user").(*store.User)

	if c.Request().Method == http.MethodPost {
		form, err := c.MultipartForm()
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_FILE")
		}
		files := form.File["files"]
		if len(files) > 10 {
			return c.String(http.StatusBadRequest, "ERR_MAX_FILES")
		}

		for _, file := range files {
			occupiedSpace, err := api.store.Files.GetOccupiedSpace(user.Id)
			if err != nil {
				return c.String(http.StatusInternalServerError, "ERR_FILE")
			}

			if (float32(file.Size)/1024)+occupiedSpace > user.Space {
				return c.String(http.StatusForbidden, "ERR_FILE_SPACE")
			}

			if file.Filename == "" {
				return c.String(http.StatusBadRequest, "ERR_FILE")
			}
			uid := uuid.New()

			src, err := file.Open()
			if err != nil {
				return c.String(http.StatusInternalServerError, "ERR_FILE")
			}
			defer src.Close()

			dest, err := os.Create("uploads/files/" + uid.String())
			if err != nil {
				os.Remove("upload/files/" + uid.String())
				return c.String(http.StatusInternalServerError, "ERR_FILE")
			}
			defer dest.Close()

			if _, err = io.CopyBuffer(dest, src, make([]byte, (1024^2)*2)); err != nil {
				os.Remove("upload/files/" + uid.String())
				return c.String(http.StatusInternalServerError, "ERR_FILE")
			}

			_, err = api.store.Files.Create(store.NewFile(uid, file.Filename, float32(file.Size)/1024, user.Id))
			if err != nil {
				return c.String(http.StatusInternalServerError, "ERR_FILE")
			}
		}
		return c.String(http.StatusOK, "OK")
	}

	if c.Request().Method == http.MethodGet {
		pageStr := c.QueryParam("page")
		sizeStr := c.QueryParam("size")
		id := c.QueryParam("id")
		if pageStr == "" && sizeStr == "" {
			if id == "" {
				return c.String(http.StatusBadRequest, "ERR_GET")
			}

			uid, err := uuid.Parse(id)
			if err != nil {
				return c.String(http.StatusBadRequest, "ERR_GET")
			}
			fileDb, err := api.store.Files.GetById(uid, user.Id)
			if err != nil {
				return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
			}

			file, err := os.Open("uploads/files/" + id)
			if err != nil {
				return c.String(http.StatusBadRequest, "ERR_GET")
			}
			defer file.Close()

			c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
			c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="`+fileDb.Name+`"`)

			_, err = io.CopyBuffer(c.Response().Writer, file, make([]byte, (1024^2)*2))
			if err != nil {
				return c.String(http.StatusBadRequest, "ERR_GET")
			}
			return c.String(http.StatusOK, "OK")
		}

		page, err := strconv.Atoi(pageStr)
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_GET")
		}
		files, err := api.store.Files.GetPage(user.Id, page, size)
		filesPublic, err := api.store.Files.GetPagePublic(page, size)
		if err != nil {
			return c.String(http.StatusInternalServerError, "ERR_GET")
		}
		*files = append(*files, *filesPublic...)
		return c.JSON(http.StatusOK, files)
	}

	if c.Request().Method == http.MethodDelete {
		id := c.QueryParam("id")
		if id == "" {
			return c.String(http.StatusBadRequest, "ERR_DEL")
		}

		uid, err := uuid.Parse(id)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_DEL")
		}
		err = api.store.Files.Delete(uid, user.Id)
		if err != nil {
			return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
		}
		err = os.Remove("uploads/files/" + id)
		if err != nil {
			return c.String(http.StatusInternalServerError, "ERR_DEL")
		}
		return c.String(http.StatusOK, "OK")
	}

	if c.Request().Method == http.MethodPut {
		id := c.FormValue("id")
		name := c.FormValue("name")
		dir_id := c.FormValue("dir")
		publicString := c.FormValue("public")
		favouriteString := c.FormValue("favourite")
		if id == "" {
			return c.String(http.StatusBadRequest, "ERR_PUT")
		}

		uid, err := uuid.Parse(id)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERR_PUT")
		}

		if name != "" {
			err = api.store.Files.UpdateField(uid, user.Id, "name", name)
			if err != nil {
				return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
			}
		}
		if dir_id != "" {
			err = api.store.Files.UpdateField(uid, user.Id, "dir", dir_id)
			if err != nil {
				return c.String(http.StatusForbidden, "ERR_FORBIDDEN")
			}
		}
		if publicString != "" {
			if publicString == "true" {
				err = api.store.Files.UpdateField(uid, user.Id, "public", true)
			}
			if publicString == "false" {
				err = api.store.Files.UpdateField(uid, user.Id, "public", false)
			}
		}
		if favouriteString != "" {
			if favouriteString == "true" {
				err = api.store.Files.UpdateField(uid, user.Id, "favourite", true)
			}
			if favouriteString == "false" {
				err = api.store.Files.UpdateField(uid, user.Id, "favourite", false)
			}
		}

		return c.String(http.StatusOK, "OK")
	}

	return c.String(http.StatusInternalServerError, "ERR_FILE")
}
