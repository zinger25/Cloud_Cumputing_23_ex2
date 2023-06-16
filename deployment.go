package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

type instance struct {
	IP string `json:"ip"`
}

func main() {
	cmd := exec.Command("bash", "first_endpoint_deployment.sh")
	output, err := cmd.Output()
	lines := strings.Split(string(output), "\n")
	if err != nil {
		log.Fatal(err)
	}
	var instances []instance
	instances = append(instances, instance{IP: lines[len(lines)-2]})

	cmd = exec.Command("bash", "sec_endpoint_deployment.sh")
	output, err = cmd.Output()
	lines = strings.Split(string(output), "\n")
	if err != nil {
		log.Fatal(err)
	}
	instances = append(instances, instance{IP: lines[len(lines)-2]})

	jsonData, err := json.Marshal(instances)
	ioutil.WriteFile("./endpoints.json", jsonData, 0644)

	cmd = exec.Command("bash", "upload_instances_data.sh", instances[0].IP, instances[1].IP)
	output, _ = cmd.Output()
	fmt.Printf(string(output))
}
