package handler

import (
	"net/http"
	"sync"

	"github.com/NandiniDhanrale/user-age-api/app"
	"github.com/gofiber/adaptor/v2"
)

var (
	once    sync.Once
	fn      http.HandlerFunc
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		a := app.New()
		fn = adaptor.FiberApp(a.Fiber)
	})
	fn(w, r)
}
