package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-tello/app/models"
	"go-tello/config"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"text/template"
)

var appContext struct {
	droneManager *models.DroneManager
}

func init() {
	appContext.droneManager = models.NewDroneManager()
}

func StartWebServer() error {
	r := gin.Default()

	/* set up router */
	// GET
	r.GET("/", viewIndexHandler)
	r.GET("/controller/", viewControllerHandler)
	// POST
	r.POST("/api/command", apiMakeHandler(apiCommandHandler))
	// File Server
	r.StaticFS("/static/", http.Dir("static"))
	// Video Streaming
	r.GET("/video/streaming", func(ctx *gin.Context) {
		appContext.droneManager.Stream.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return r.Run(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port))
}

var apiValidPath = regexp.MustCompile("^/api/(command|video)")

func apiMakeHandler(fn func(ctx *gin.Context)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		matches := apiValidPath.FindStringSubmatch(ctx.Request.URL.Path)
		if len(matches) <= 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "invalid path"})
			return
		}

		ctx.Next()
		fn(ctx)
	}
}

func apiCommandHandler(ctx *gin.Context) {
	command, ok := ctx.GetPostForm("command")
	if !ok || command == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid api request"})
		return
	}

	log.Printf("action=apiCommandHandler command=%s", command)

	var err error = nil

	switch command {
	case "ceaseRotation":
		appContext.droneManager.CeaseRotation()
	case "takeOff":
		err = appContext.droneManager.TakeOff()
	case "land":
		err = appContext.droneManager.Land()
	case "hover":
		appContext.droneManager.Hover()
	case "up":
		err = appContext.droneManager.Up(appContext.droneManager.Speed)
	case "down":
		err = appContext.droneManager.Down(appContext.droneManager.Speed)
	case "clockwise":
		err = appContext.droneManager.Clockwise(appContext.droneManager.Speed)
	case "counterClockwise":
		err = appContext.droneManager.CounterClockwise(appContext.droneManager.Speed)
	case "forward":
		err = appContext.droneManager.Forward(appContext.droneManager.Speed)
	case "left":
		err = appContext.droneManager.Left(appContext.droneManager.Speed)
	case "right":
		err = appContext.droneManager.Right(appContext.droneManager.Speed)
	case "backward":
		err = appContext.droneManager.Backward(appContext.droneManager.Speed)
	case "speed":
		appContext.droneManager.Speed = getSpeed(ctx)
	case "frontFlip":
		err = appContext.droneManager.FrontFlip()
	case "backFlip":
		err = appContext.droneManager.BackFlip()
	case "leftFlip":
		err = appContext.droneManager.LeftFlip()
	case "rightFlip":
		err = appContext.droneManager.RightFlip()
	case "throwTakeOff":
		err = appContext.droneManager.ThrowTakeOff()
	default:
		err = errors.New("command not found")
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.Status(http.StatusOK)
}

func getSpeed(ctx *gin.Context) int {
	speed, ok := ctx.GetPostForm("speed")
	if !ok {
		return models.DefaultSpeed
	}

	newSpeed, err := strconv.Atoi(speed)
	if err != nil {
		return models.DefaultSpeed
	}

	return newSpeed
}

func getTemplate(temp string) (*template.Template, error) {
	return template.ParseFiles("app/views/layout.html", temp)
}

func viewIndexHandler(ctx *gin.Context) {
	t, err := getTemplate("app/views/index.html")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = t.Execute(ctx.Writer, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.Status(http.StatusOK)
}

func viewControllerHandler(ctx *gin.Context) {
	t, err := getTemplate("app/views/controller.html")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = t.Execute(ctx.Writer, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.Status(http.StatusOK)
}
