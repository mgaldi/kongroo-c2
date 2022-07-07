# Notes

This repo doesn't follow the main development, and it's here only because I'm looking for jobs it's listed in the CV. This repo has only some basic functionalities but if you want to contribute on the private repo or wish to participate in any way, get in contact with me. 

# KONGROO C2

This project is a personal exercise to get a more in-depth understanding of a C2 and the various ways implants work and interact with it.
The project is far from a release stage but its core functionalities are working.

Quick info on folders:

- /agent: contains the files that create the binary for the agent.

- /c2: contains the handlers and the back infrastracture for the callback communication with the agents.

- /web: for now it only contains the src for the web interface.

## Test C2
`docker run --name mongodb -d -p 27017:27017 mongo`


`docker run --name redis -d -p 6379:6379 redis`

```
cd c2
go run main.go
```
