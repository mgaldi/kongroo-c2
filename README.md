# KONGROO C2

This project is a personal exercise to get a more in-depth understanding of a C2 and the various ways implants work and interact.
The project is far from a release stage but its core functionalities are working.

Quick info on folders:

- /agent: contains the files that create the binary for the agent.

- /c2: contains the handlers and the back infrastracture for the communication with the agents.

- /web: web interface to interact with agents and eventually generate payloads

The project is being currently developed on gitlab.com/mgdi/kongroo-c2

## Test C2
`docker run --name mongodb -d -p 27017:27017 mongo`


`docker run --name redis -d -p 6379:6379 redis`

```
cd c2
go run main.go
```
