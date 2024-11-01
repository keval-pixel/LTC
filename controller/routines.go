package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

func cli(cmd_chan chan string, ch_controller chan int) {
	reader := bufio.NewReader(os.Stdin)
	input := ""
	for {
		//fmt.Println("waiting for input controller")
		<-ch_controller
		//fmt.Println("After input controller")
		for input == "" {
			color.New(color.FgHiMagenta).Printf("# ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
		}
		cmd_chan <- input
		if input == "exit" {
			return
		}
		input = ""
	}
}

func test_cli(bash string, cmd_chan chan string, start_control_ch chan int, stop_control_ch chan error) {
	reader := bufio.NewReader(os.Stdin)
	input := ""
	running := true
	first := true

	for running {
		if first {
			go func() {
				<-stop_control_ch
				running = false
				return
			}()
			first = false
		}
		<-start_control_ch
		for input == "" && running {
			color.New(color.FgHiCyan).Printf("%s ", bash)
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
		}
		cmd_chan <- input
		input = ""
	}

	return
}

func handleConnection(listener net.Listener) net.Conn {

	conn, err := listener.Accept()
	if err != nil {
		color.Red("Error accepting connection:", err)
		return nil
	}
	//connection = true
	color.Green("Feedback plugin connected")

	return conn
}

func runCommand(filename string, name string, arg ...string) (cmd *exec.Cmd) {
	//Wg.Add(1)
	//defer Wg.Done()
	var output *os.File
	var errors *os.File

	cmd = exec.Command(name, arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		color.Red("Error creating stdout pipe:", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		color.Red("Error creating stderr pipe:", err)
		return
	}
	if filename == "" {
		output = os.Stdout
		errors = os.Stderr
	} else {
		outfile, err := os.Create(filename + "_stdout")
		errfile, err := os.Create(filename + "_stderr")
		if err != nil {
			color.Red("Error creating file:", err)
			return
		}

		output = outfile
		errors = errfile
	}
	go func() {
		if _, err := io.Copy(output, stdout); err != nil {
			color.Red("Error copying stdout: %v", err)
		}
	}()
	go func() {
		if _, err := io.Copy(errors, stderr); err != nil {
			color.Red("Error copying stderr: %v", err)
		}
	}()

	return
}

func collect_metrics(metric string, metric_percentage float64, feedback_time int, conn net.Conn, deployment string, ch_result chan error) {

	var msg Message
	var msg_arr = make([]Message, feedback_time)
	var load_unit int32
	var margin float64
	var lower_margin float64
	var upper_margin float64
	var step int32
	data_ch := make(chan Message)
	running := true
	//give some time for the stress tests to start

	fmt.Print("\r\033[K")
	for i := 3; i >= 1; i-- {
		fmt.Printf("\rStarting in %d", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Print("\r\033[K")
	//send a signal to the client to start sending the metrics
	_, err := conn.Write([]byte{0x11})

	if err != nil {
		color.Red("Error sending: %v", err)
		return
	}

	margin = min(max(metric_percentage*0.1, 1), 5)
	lower_margin = (metric_percentage - margin)
	upper_margin = (metric_percentage + margin)

	//collect metrics from the metrics agent
	go func() {
		decoder := json.NewDecoder(conn)
		for {
			if err := decoder.Decode(&msg); err != nil {
				color.Red("Error decoding JSON: %v", err)
				return
			}
			data_ch <- msg
		}
	}()

	if deployment != "" {
		load_unit = 1
	} else {
		load_unit = get_vus()
	}

	i := 1
	for {
		select {
		case <-ch_result:
			if running {
				_, err := conn.Write([]byte{0x11})
				if err != nil {
					color.Red("Error sending: %v", err)
					return
				}
			}
			running = false
			return
		case msg = <-data_ch:
			msg_arr[i-1] = msg
			if i%feedback_time == 0 {
				metrics_avg := calculateAverage(msg_arr[:], metric)
				fmt.Printf("Average %s value is: %.1f%%\n", metric, metrics_avg)
				if metrics_avg < lower_margin {

					if deployment != "" { //if the load test is running in kubernetes
						load_unit = get_replicas(deployment)
						step = compute_step(metric_percentage, metrics_avg, load_unit)
						load_unit += step
						color.Yellow("Increasing number of replicas by %d\n", step)
						update_replicas(load_unit, deployment)
					} else {
						load_unit = get_vus()
						step = compute_step(metric_percentage, metrics_avg, load_unit)
						load_unit += step
						color.Yellow("Increasing number of users by %d\n", step)
						k6_update_user(load_unit)
					}
				} else if metrics_avg > upper_margin {
					if load_unit > 0 {
						if deployment != "" { //if the load test is running in kubernetes
							load_unit = get_replicas(deployment)
							step = compute_step(metric_percentage, metrics_avg, load_unit)
							load_unit -= step
							color.Yellow("decreasing number of replicas by %d\n", step)
							update_replicas(load_unit, deployment)
						} else {
							load_unit = get_vus()
							step = compute_step(metric_percentage, metrics_avg, load_unit)
							load_unit -= step
							color.Yellow("decreasing number of users by %d\n", step)
							k6_update_user(load_unit)
						}
					} else {
						color.Yellow("%s utilization is above %.1f%% without load testing\n", metric, metric_percentage)
					}
				} else {
					color.HiYellow("%s utilization beween %.1f%% and %.1f%%\n", metric, lower_margin, upper_margin)
					if deployment != "" {
						color.HiYellow("Number of replicas is %d\n", load_unit)
					} else {
						color.HiYellow("Number of users is %d\n", load_unit)
					}
				}
				if load_unit == -1 {
					load_unit = 1 //reset load_unit in case the deployment does not exist anymore or an error occurs
				}
				i = 0
			}
			i++
		}
	}
}

func calculateAverage(samples []Message, metric string) float64 {
	var total float64
	for _, sample := range samples {
		if metric == "cpu" {
			total += sample.CPU
		} else if metric == "memory" {
			total += sample.Memory
		} else {
			color.Red("Error: invalid metric defined")
		}

	}
	return total / float64(len(samples))
}

func get_key_value(data string, jsonData []byte) int32 {
	var result map[string]interface{}

	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		color.Red("Error unmarshalling JSON: %v", err)
		return -1
	}
	// Access nested values with type assertions
	jData := result["data"].(map[string]interface{})
	attributes := jData["attributes"].(map[string]interface{})
	value := attributes[data].(float64)

	return int32(value)
}

func compute_step(metrics_percentage float64, metrics_avg float64, load_unit int32) int32 {
	difference := math.Abs(metrics_percentage - metrics_avg)
	load_per_unit := metrics_avg / float64(load_unit)
	step := min(max(int32(difference/load_per_unit), 1), 5)

	return step
}
