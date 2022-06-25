package router

import (
	//"fmt"

	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"gitlab.com/mgdi/kongroo-c2/c2/config"
	mongo "gitlab.com/mgdi/kongroo-c2/c2/database/mongo"
	"gitlab.com/mgdi/kongroo-c2/c2/database/redis"
	"gitlab.com/mgdi/kongroo-c2/c2/helpers"
	"gitlab.com/mgdi/kongroo-c2/c2/router/handlers"
)

func Run() {

	handlers.InitHub()

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Start a go routine needed for swagger to generate doc
	r.Mount("/swagger", httpSwagger.WrapHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// r.Get("/testWs", handlers.TestWs)
	// r.Get("/test2Ws", handlers.Test2Ws)

	r.Get("/ws", handlers.WebSocket)

	// Routes that handle agents registration

	/////////////////////////////
	///////          REGISTRATION
	/////////////////////////////

	r.Route("/reg/{agent}", func(r chi.Router) {
		r.Use(agentCtx)
		// Get info about agent
		r.Get("/", handlers.GetAgent)
		// Create a new agent
		r.Post("/", handlers.CreateAgent)

	})

	/////////////////////////////
	///////                  TASK
	/////////////////////////////

	r.Route("/tasks/{agent}", func(r chi.Router) {
		r.Use(taskCtx)
		r.Get("/", handlers.GetTask)

		// to construct a posttaskresult request you have to provide base64(command)+base64(giannino)+base64(output)
		r.Post("/", handlers.PostTaskResult)

	})
	r.Route("/tasks/{agent}/{task}", func(r chi.Router) {
		r.Use(createTaskCtx)
		r.Post("/", handlers.CreateTask)
	})

	r.Route("/tasks/{agent}/history", func(r chi.Router) {
		r.Use(historyCtx)
		r.Get("/", handlers.GetCommandHistory)
	})

	/////////////////////////////
	///////                 AGENT
	/////////////////////////////
	r.Route("/agents/getall", func(r chi.Router) {
		r.Get("/", handlers.GetAllAgents)
	})

	r.Route("/agents/getallbase", func(r chi.Router) {
		r.Get("/", handlers.GetAllAgentsBase)
	})

	handler := cors.Default().Handler(r)

	http.ListenAndServe(config.Configs["c2.port"], handler)
}

func historyCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentName := chi.URLParam(r, "agent")
		ctx := context.WithValue(r.Context(), "agent", agentName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func taskCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context

		agentName := chi.URLParam(r, "agent")

		if r.Method == "GET" {

			// Check if agent exists
			if found := existCheck(agentName, w); !found {
				helpers.WriteResponse(w, 1, errors.New("Agent not found"))
				return
			}

			// Get task from Redis Server
			task, _ := redis.RedisCl.Get(agentName)

			ctx = context.WithValue(r.Context(), "task", task)
		}

		if r.Method == "POST" {
			ctx = context.WithValue(r.Context(), "agent", agentName)

			if found := existCheck(agentName, w); !found {
				helpers.WriteResponse(w, 1, errors.New("Agent not found"))
				return
			}

			ctx = context.WithValue(r.Context(), "agent", agentName)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createTaskCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentName := chi.URLParam(r, "agent")
		newTask := chi.URLParam(r, "task")

		if found := existCheck(agentName, w); !found {
			helpers.WriteResponse(w, 1, errors.New("Agent not found"))
			return
		}
		agentTask := [2]string{agentName, newTask}
		ctx := context.WithValue(r.Context(), "agentTask", agentTask)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func agentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentName := chi.URLParam(r, "agent")
		var ctx context.Context

		if r.Method == "POST" {
			// Check if agent is alreagy registered
			_, err := mongo.MongoCl.GetAgent(agentName)
			if err == nil {
				helpers.WriteResponse(w, 1, errors.New("Agent is already registered"))
				return
			}
			// Insert agentName inside context
			ctx = context.WithValue(r.Context(), "id", agentName)
		}

		if r.Method == "GET" {
			// Check if agent is alreagy registered
			agent, err := mongo.MongoCl.GetAgent(agentName)
			if err != nil {
				helpers.WriteResponse(w, 1, err)
				return
			}
			// Insert agentName inside context
			ctx = context.WithValue(r.Context(), "agent", agent)
		}
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func existCheck(agent string, w http.ResponseWriter) bool {
	if _, err := mongo.MongoCl.GetAgent(agent); err != nil {
		return false
	}
	return true
}
