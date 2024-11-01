package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func welcome_message() {
	color.HiMagenta("Welcome to LTC - Load Testing Controller\n\n")
	println("Please select on of the following options:\n1.Connect the metrics agent\n2.Start load testing\n3.Exit\n")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func emptyChannel(ch chan int) {
	for {
		select {
		case <-ch:
			// Continue reading from the channel
		default:
			// Break the loop if the channel is empty
			return
		}
	}
}

func get_testplan(reader *bufio.Reader) (string, string) {

	input := ""
	workingDir := ""
	homefile, err := filepath.Abs(filepath.Dir(os.Args[0])) //executable file
	homeDir := filepath.Dir(homefile)                       //directory where executable is located

	if err != nil {
		log.Fatal(err)
	}
	script := "test.js"

	for !dirExists(workingDir) {
		fmt.Printf("Enter working directory [default:%s]: ", homeDir)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			workingDir = homeDir
		} else {
			workingDir = input
		}
		if !dirExists(workingDir) {
			color.Red("Error: Directory does not exist")
		}
	}

	scriptPath := filepath.Join(workingDir, script)

	for {
		fmt.Printf("Enter script path [default: %s]: ", scriptPath)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if fileExists(input) {
			scriptPath = input
			break
		} else if input == "" {
			if fileExists(scriptPath) {
				break
			} else {
				color.Red("Default path does not exist, specify a script path")
			}
		} else if fileExists(filepath.Join(workingDir, input)) {
			scriptPath = filepath.Join(workingDir, input)
			break
		} else if input == "exit" {
			return "", ""
		} else {
			color.Red("File not found! Try again")
			continue
		}
	}

	// Command to run k6
	color.Yellow("full path set to: %s", scriptPath)

	return scriptPath, workingDir
}

func read(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)

	return input
}

func dockerImageExists(container_image string, tag string) bool {
	imageURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s", container_image, tag)
	resp, err := http.Get(imageURL)
	if err != nil {
		color.Red("Error retrieving container image from Docker Hub.")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		color.Red("container image does not exist on Docker Hub.")
		return false
	}

	if resp.StatusCode != 200 {
		color.Red("failed to fetch Docker Hub data: %s", resp.Status)
		return false
	}

	return true
}

func executableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}