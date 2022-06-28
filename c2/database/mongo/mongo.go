package database

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MongoCl *Client

// Client : MongoDB data for the current session and database/collection
type Client struct {
	Client   *mongo.Client
	Context  context.Context
	Database *mongo.Database
}

type AgentBaseInfo struct {
	PCID      string `bson:"pcid,omitempty" json:"pcid,omitempty"`
	Name      string `bson:"name,omitempty" json:"name,omitempty"`
	IP        string `bson:"ip,omitempty" json:"ip,omitempty"`
	Platform  string `bson:"platform,omitempty" json:"platform,omitempty"`
	Timestamp string `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
}
type AgentInfo struct {
	PCID     string    `bson:"pcid,omitempty" json:"pcid,omitempty"`
	IP       string    `bson:"ip,omitempty" json:"ip,omitempty"`
	Platform string    `bson:"platform,omitempty" json:"platform,omitempty"`
	Command  string    `bson:"command,omitempty" json:"command,omitempty"`
	Output   string    `bson:"output,omitempty" json:"output,omitempty"`
	Date     time.Time `bson:"date,omitempty" json:"date,omitempty"`
}

type Command struct {
	Command string
	Output  string
}
type Commands []Command

// NewClient : Init connection and return data
func NewClient(connectionString string, database string) (err error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		return errors.Errorf("Could not create a new Mongo Client: %s", err.Error())
	}

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		return errors.Errorf("Could not connect to mongo: %s", err.Error())
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return errors.Errorf("Could not ping mongo: %s", err.Error())
	}

	MongoCl = &Client{
		Client:   client,
		Context:  ctx,
		Database: client.Database(database),
	}
	return
}

//GetCommandHistory: Returns commands history for a givn agent
func (db *Client) GetCommandHistory(col string) (results []bson.M, err error) {
	found, err := db.FindCollection(col)
	if !found {
		log.Println(err)
		return results, errors.New("No agent found")
	}
	if err != nil {
		log.Println(err)
		return results, errors.New("Error connecting to DB")
	}

	opts := options.Find()
	opts.SetProjection(bson.D{
		{"command", 1},
		{"output", 1},
		{"_id", 0}})

	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	// count, err := db.Database.Collection(col).EstimatedDocumentCount(ctx)
	if err != nil {
		log.Fatal("Mongo error while counting documents for collection" + col)
	}
	// skip := count - 3
	// if count < 3 {
	// 	skip = 0
	// }
	// opts.SetSkip(skip)
	defer cancel()
	cursor, err := db.Database.Collection(col).Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println(err)
		return results, errors.New("Error fetching command history")
	}
	if err = cursor.All(db.Context, &results); err != nil {
		log.Println("Error stuff")
		return results, errors.New("Error parsing command history")
	}

	return results, nil
}

// CloseClient : Close connection
func (db *Client) CloseClient() {
	db.Client.Disconnect(db.Context)
}

//CreateAllAgentsCollection: Creates the collection Agents
func (db *Client) CreateAllAgentsCollection() error {
	if found, err := db.FindCollection("Agents"); found {
		return nil
	} else if err != nil {
		log.Fatal(err)
	}
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"name"},
		"properties": bson.M{
			"pcid": bson.M{
				"bsonType":    "string",
				"description": "the id of the agent",
			},
			"name": bson.M{
				"bsonType":    "string",
				"description": "the name of the agent",
			},
			"ip": bson.M{
				"bsonType":    "string",
				"description": "ip address of agent",
			},
			"platform": bson.M{
				"bsonType":    "string",
				"description": "platform of agent",
			},
			"timestamp": bson.M{
				"bsonType":    "string",
				"description": "the timestamp the agent was first run",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()

	if err := db.Database.CreateCollection(ctx, "Agents", opts); err != nil {
		return errors.Errorf("Could not create collection: %s", err)
	}
	return nil
}

// CreateAgentCollection : create agent
func (db *Client) CreateAgentCollection(col string) error {
	if found, err := db.FindCollection(col); found {
		return nil
	} else if err != nil {
		log.Fatal(err)
	}

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"date", "command", "output", "ip"},
		"properties": bson.M{
			"date": bson.M{
				"bsonType":    "date",
				"description": "the time of the current command, which is required and must be a int64",
			},
			"command": bson.M{
				"bsonType":    "string",
				"description": "command launched",
			},
			"output": bson.M{
				"bsonType":    "string",
				"description": "output of command",
			},
			"ip": bson.M{
				"bsonType":    "string",
				"description": "ip address of agent",
			},
			"platform": bson.M{
				"bsonType":    "string",
				"description": "platform of agent",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()

	if err := db.Database.CreateCollection(ctx, col, opts); err != nil {
		return errors.Errorf("Could not create collection: %s", err)
	}
	return nil
}

// InsertRow : insert row
func (db *Client) InsertAgentRow(col string, data interface{}) error {
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()
	_, err := db.Database.Collection(col).InsertOne(ctx, data)
	if err != nil {
		return errors.Errorf("Could not insert data: %s", err.Error())
	}
	//res.InsertedID
	log.Println("INSERITA")
	return nil
}

//InsertAgentBaseRow: Insert base info in collection Agents
func (db *Client) InsertAgentBaseRow(data interface{}) error {
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()
	_, err := db.Database.Collection("Agents").InsertOne(ctx, data)
	if err != nil {
		return errors.Errorf("Could not insert data: %s", err.Error())
	}
	//res.InsertedID
	log.Println("INSERITA BASE")
	return nil
}

//GetAgentsBase: Get base info for all agents
func (db *Client) GetAgentsBase() (results []bson.M, err error) {
	opts := options.Find()
	opts.SetProjection(bson.D{
		{"pcid", 1},
		{"name", 1},
		{"ip", 1},
		{"platform", 1},
		{"_id", 0}})
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()
	cursor, err := db.Database.Collection("Agents").Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println(err)
		return results, errors.New("Error fetching all agents base information")
	}
	if err = cursor.All(db.Context, &results); err != nil {
		log.Println("Error stuff")
		return results, errors.New("Error parsing all agents base information")
	}

	return results, nil
}

//GetAgentsBase: Returns all the collections' names
//Note: after commit 6946ef71 this function also returns the collection "Agents"
func (db *Client) ListAllAgents() ([]string, error) {
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()

	result, err := db.Database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, errors.Errorf("Could not list collection: %s", err)
	}
	return result, nil
}

// FindCollection : check all collections and return true if the collection passed as arg exists
func (db *Client) FindCollection(col string) (bool, error) {
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()
	result, err := db.Database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, errors.Errorf("Could not list collection: %s", err)
	}
	// log.Println(result)
	for _, collection := range result {
		if collection == col {
			// fmt.Println(collection, col)
			return true, nil
		}
	}
	return false, nil
}

//GetAgent: Returns info for the given agent collection
func (db *Client) GetAgent(agent string) (*AgentInfo, error) {
	ctx, cancel := context.WithTimeout(db.Context, time.Second*10)
	defer cancel()
	opts := options.FindOne()
	opts.SetProjection(bson.D{
		{"name", 1},
		{"ip", 1},
		{"platform", 1},
		{"date", 1},
		{"_id", 0}})

	result := db.Database.Collection(agent).FindOne(ctx, bson.D{}, opts)
	if result.Err() != nil {
		return &AgentInfo{}, errors.Errorf("Could not list collection: %s", result.Err().Error())
	}

	var agentInfo *AgentInfo
	err := result.Decode(&agentInfo)
	if err != nil {
		return &AgentInfo{}, errors.Errorf("Could not list collection: %s", result.Err().Error())
	}

	return agentInfo, nil
}

func (db *Client) IsAgent(agent string) error {
	if found, err := db.FindCollection(agent); found {
		return nil
	} else if err != nil {
		return err
	}
	return errors.New("No agent")
}
