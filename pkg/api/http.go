package api

import (
	"context"
	"github.com/gin-gonic/gin"
	pkgLog "github.com/kapitanov/tg-waqi-bot/pkg/log"
	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"time"
)

var log = pkgLog.New("api")

// Server wraps WebAPI logic
type Server interface {
	// Start start WebAPI server
	Start()

	// Stop stops WebAPI server
	Stop()
}

// NewServer configures new WebAPI server instance
func NewServer(service waqi.Service, listenAddr string) (Server, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	http.Handle("/", router)

	// REST API
	controller := &restController{service: service}
	router.GET("/api/status/geo", controller.GetByGeo)
	router.GET("/api/status/city/:city", controller.GetByCity)
	router.GET("/api/status/station/:id", controller.GetByStation)

	// Static files
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		return nil, err
	}
	router.NoRoute(func(c *gin.Context) {
		dir, file := path.Split(c.Request.RequestURI)
		ext := filepath.Ext(file)
		if file == "" || ext == "" {
			c.File("./www/index.html")
		} else {
			c.File("./www" + path.Join(dir, file))
		}
	})

	s := &server{
		listenAddr: listenAddr,
		done:       make(chan bool),
	}
	return s, nil
}

type server struct {
	listenAddr string
	httpServer *http.Server
	done       chan bool
}

// Start start WebAPI server
func (s *server) Start() {
	s.httpServer = &http.Server{Addr: s.listenAddr}

	go func() {
		log.Printf("listening on \"%s\"\n", s.httpServer.Addr)

		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on \"%s\": %v\n", s.httpServer.Addr, err)
		}

		s.done <- true
	}()
}

// Stop stops WebAPI server
func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.httpServer.SetKeepAlivesEnabled(false)
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Fatalf("could not gracefully shutdown the server: %v\n", err)
	}

	<-s.done
}
