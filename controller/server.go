package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type Message struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

func main() {

	welcome_message()
	var choice string
	var conn net.Conn
	var err error
	var listener net.Listener
	connected := false
	cmdChan := make(chan string)
	cmd_ch := make(chan string)
	ch_results := make(chan error)
	stop_cli := make(chan error)
	inputcontroller := make(chan int)
	bash_control := make(chan int)
	var cmd *exec.Cmd
	var deployment_name = ""
	var input string
	var tool, container_image string
	var metrics_threshold float64
	var metric string
	var feed_time int
	test_run := false
	reader := bufio.NewReader(os.Stdin)

	go cli(cmdChan, inputcontroller)

	// Accept connections and handle them in a new goroutine

	running := true
	for running {
		inputcontroller <- 1
		choice = <-cmdChan
		choice = strings.TrimSpace(choice)
		if choice == "1" {
			if connected {
				for {
					color.Yellow("There is a connection an ongoing connection, do you want to close it?[Y/n]")
					input = read(reader)

					if input == "y" || input == "" {
						conn.Close()
						connected = false
						break
					} else if input == "n" {
						break
					} else {
						continue
					}
				}

			} else {
				listener, err = net.Listen("tcp", "0.0.0.0:8080")
				if err != nil {
					fmt.Printf("Error listening: %v\n", err)
					continue
				}
				//color.Yellow("Server listening on 0.0.0.0:8080")
			}
			// Handle the connection in a new goroutine
			if !connected {
				println("waiting for connection...")
				conn = handleConnection(listener)
				if conn == nil {
					color.Red("Could not connect to metrics agent!")
					continue
				}
				connected = true
			}
		} else if choice == "2" {
			if test_run {
				color.Yellow("Test already running... skipping")
				continue
			}

			if connected {
				for {
					fmt.Printf("Select the metric to be monitored on the target server [CPU/Memory]: ")
					input, _ = reader.ReadString('\n')
					input = strings.TrimSpace(input)

					if input == "cpu" || input == "memory" {
						metric = input
						break
					} else {
						color.Red("invalid metric entered! choose CPU or Memory.")
					}
				}
				for {
					fmt.Printf("Select the target %s utilization [1-100]%%: ", metric)
					input, _ = reader.ReadString('\n')
					input = strings.TrimSpace(input)

					metrics_threshold, err = strconv.ParseFloat(input, 32)

					if metrics_threshold >= 1 && metrics_threshold <= 100 {
						break
					} else {
						color.Red("invalid value entered! choose between 1 and 100.")
					}
				}
				for {
					fmt.Printf("Select the feedback time in seconds [10-60]: ")
					input, _ = reader.ReadString('\n')
					input = strings.TrimSpace(input)

					feed_time, err = strconv.Atoi(input)

					if feed_time >= 10 && feed_time <= 60 {
						break
					} else {
						color.Red("invalid feedback time entered! choose between 10 and 60 seconds.")
					}
				}
			}

			for {
				fmt.Print("Run on Kubernetes or K6 [default: K6]: ")
				input = read(reader)

				if input == "kubernetes" {
					kubeconfig := filepath.Join(homeDir(), ".kube", "config")
					if fileExists(kubeconfig) {
						break
					} else {
						color.Red("Kubernetes admin configurations not found on %s\n", kubeconfig)
						input = ""
						for input != "exit" {
							fmt.Print("Enter the path for the kubernetes admin configuration:")
							input = read(reader)
							if fileExists(input) {
								kubeconfig = input
								break
							} else if input == "" || input == "exit" {
								continue
							} else {
								color.Red("Error: File not found")
							}
						}
					}
				} else if input == "k6" || input == "" {
					break
				} else {
					color.Red("Wrong input entered! choose between Kubernetes or K6.")
				}
			}

			if input == "kubernetes" {

				var checks bool
				tool = ""

				for tool == "" {
					fmt.Print("Provide a deployment name: ")
					tool = read(reader)
				}

				for {
					fmt.Print("Provide the container image: ")
					container_image = read(reader)

					substrings := strings.Split(container_image, ":")
					size := len(substrings)

					if size == 1 {
						checks = dockerImageExists(substrings[0], "latest")
					} else if size == 2 {
						checks = dockerImageExists(substrings[0], substrings[1])
					} else {
						color.Red("Wrong container image name format!")
						continue
					}

					if checks {
						break
					} else {
						continue
					}

				}

				// Start the load tests
				fmt.Println("Press enter to start the test")
				_, _ = reader.ReadString('\n')

				deployment_name = deploy_replicas(1, tool, container_image)
				if deployment_name == "" {
					continue
				}
				if connected {
					go collect_metrics(metric, metrics_threshold, feed_time, conn, deployment_name, ch_results)
				}
				go test_cli("cluster>", cmd_ch, bash_control, stop_cli)

				color.Yellow("\nStarting %s cluster Test. Available commands are:\n1.start\n2.status <optional: pod_name>\n3.increase <optional: increase_step>\n4.decrease <optional: decrease_step>\n5.stop\n6.exit\n\n", tool)

				test_run = true

			load_controller:
				for {
					bash_control <- 1
					command := <-cmd_ch
					options := strings.Fields(command)
					if len(options) <= 1 {
						switch command {
						case "start", "1":
							if test_run == false {
								deployment_name = deploy_replicas(1, tool, container_image)
								if deployment_name == "" {
									continue
								}
								if deployment_name != "" && connected == true {
									go collect_metrics(metric, metrics_threshold, feed_time, conn, deployment_name, ch_results)
								}
								test_run = true
							} else {
								color.Yellow("Deployments already running.")
							}
							break
						case "status", "2":
							get_pods(deployment_name)

						case "increase", "3":
							if test_run == false {
								color.Yellow("Load tests are not running.")
							} else {
								aux := get_replicas(deployment_name) + 1
								update_replicas(aux, deployment_name)
							}
						case "decrease", "4":
							if test_run == false {
								color.Yellow("Load tests are not running.")
							} else {
								aux := get_replicas(deployment_name) - 1
								update_replicas(aux, deployment_name)
							}

						case "stop", "5":
							if test_run == false {
								color.Yellow("Load tests are not running.")
							} else {
								fmt.Println("stopping test deployment...")
								err = delete_deployment(deployment_name)
								if connected {
									ch_results <- err
								}
								test_run = false
							}
						case "exit", "6":
							if test_run != false {
								color.Yellow("Please stop the deployments first.")
							} else {
								stop_cli <- nil
								deployment_name = ""
								tool = ""
								break load_controller
							}

						case "":
							continue

						default:
							color.Red("Invalid command.")
							continue
						}
					} else if len(options) == 2 {
						switch options[0] {
						case "status", "2":
							get_pod_logs(options[1])
						case "increase", "3":
							if test_run == false {
								color.Yellow("Load tests are not running")
							} else {
								step, err := strconv.Atoi(options[1])
								if err != nil {
									color.Red("Error: %v\n", err)
									continue
								}
								aux := get_replicas(deployment_name) + int32(step)
								update_replicas(aux, deployment_name)
							}
						case "decrease", "4":
							if test_run == false {
								color.Yellow("Load tests are not running")
							} else {
								step, err := strconv.Atoi(options[1])
								if err != nil {
									color.Red("Error: %v\n", err)
									continue
								}
								aux := get_replicas(deployment_name) - int32(step)
								update_replicas(aux, deployment_name)
							}
						default:
							color.Red("Invalid command.\n")
							continue
						}
					} else {
						color.Red("Invalid command %s.\n", command)
						continue
					}
				}

			} else {
				if !executableExists("k6") {
					color.Red("K6 is not installed on the machine")
					continue
				}
				scriptPath, workingDir := get_testplan(reader)
				if scriptPath == "" {
					continue
				}
				report := filepath.Join(workingDir, "k6_report")
				cmd = runCommand(report, "k6", "run", "--address", "0.0.0.0:6565", scriptPath)

				// Start the load tests
				fmt.Println("Press enter to start the test")
				_, _ = reader.ReadString('\n')

				test_run = true

				//allow metrics to take decisions of virtual users
				if connected {
					go collect_metrics(metric, metrics_threshold, feed_time, conn, deployment_name, ch_results)
				}

				go func() {
					fmt.Println("Starting K6 test...")
					err = cmd.Run()
					if connected {
						ch_results <- err
					}
					stop_cli <- nil
					test_run = false
					color.Green("K6 test finished!")
					color.Yellow("Press enter to continue...")
				}()

				//place a k6 cli to take input from the users
				go test_cli("k6>", cmd_ch, bash_control, stop_cli)

				color.Yellow("\nStarting K6 Test. Available commands are:\n1.stop\n2.continue\n3.increase <optional: increase_step>\n4.decrease optional: decrease_step\n5.status\n\n")
				// Wait for the command to finish
				for test_run {
					bash_control <- 1
					command := <-cmd_ch
					options := strings.Fields(command)
					if len(options) <= 1 {
						switch command {
						case "stop", "exit", "1":
							color.Yellow("Stopping K6 Tests...")
							k6_stop()

						case "continue", "2":

							color.Yellow("Resuming K6 Tests...")
							k6_continue()

						case "increase", "3":
							color.Yellow("Increasing 1 VU...")
							k6_update_user(get_vus() + 1)

						case "decrease", "4":
							color.Yellow("Decreasing 1 VU...")
							k6_update_user(get_vus() - 1)

						case "status", "5":
							k6_get_status()

						case "":
							continue
						default:
							color.Red("invalid k6 command")
							continue
						}
					} else if len(options) == 2 {
						step, err := strconv.Atoi(options[1])
						if err != nil {
							color.Red("Error: %v\n", err)
							continue
						}
						switch options[0] {
						case "increase", "3":
							color.Yellow("Increasing VU by %d...", int32(step))
							k6_update_user(get_vus() + int32(step))

						case "decrease", "4":
							color.Yellow("Decreasing VU by %d...", int32(step))
							k6_update_user(get_vus() - int32(step))
						default:
							color.Red("Invalid command.\n")
							continue
						}
					} else {
						color.Red("Invalid command.\n")
						continue
					}

				}
			}

		} else if choice == "exit" || choice == "3" {
			color.Yellow("Exiting...")
			running = false
			if connected {
				listener.Close()
				conn.Close()
			}
			close(cmd_ch)
			close(inputcontroller)
			close(bash_control)
			close(stop_cli)
			close(cmdChan)
			close(ch_results)
		} else {
			color.Red("Wrong option, select again")
		}
	}
	color.Green("Program exited.")
}
