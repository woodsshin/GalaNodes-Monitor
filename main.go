// Copyright 2020-present woodsshin. All rights reserved.
// Use of this source code is governed by GNU General Public License v2.0.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"galamonitor/galanodes"
)

var (
	clear map[string]func() //create a map for storing clear funcs

	wg sync.WaitGroup
)

func Init() {

	galanodes.Init()

	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	fmt.Println("*** Welcome To Gala node monitor ***")

	printHelp()
}

func main() {

	Init()

	go runConsole()

	go galanodes.RegularReport()

	for {
		galanodes.QueryAllNodes(true)

		var duration time.Duration
		duration = time.Duration(galanodes.GetMonitorInterval())
		time.Sleep(duration * time.Second)
	}
}

func clearScreen() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func printHelp() {
	fmt.Println("Type your shell command and enter.")
	fmt.Println("")
	fmt.Println("* Command lists")
	fmt.Println("help : prints available command lists")
	fmt.Println("")
	fmt.Println("discord : discord webhook test")
	fmt.Println("nodes : prints the current status of nodes")
	fmt.Println("find [key] : find nodes which have a key in name, address, machine ID.")
	fmt.Println("save : output nodes information to output.csv")
	fmt.Println("cls : clear screen")
	fmt.Println("")
	fmt.Println("stats : run \"gala-node stats\" on all nodes")
	fmt.Println("restart [index|address|all] : run \"sudo systemctl restart gala-node\" on a gala node or all nodes. The index is non zero based.")
	fmt.Println("reboot [index|address|all] : reboot operating system of a node or all nodes. The index is non zero based.")
	fmt.Println("")
	fmt.Println("exit : exit program")
	fmt.Println("")
}

func runConsole() {
	scanner := bufio.NewScanner(os.Stdin)

	var (
		cmd   string
		parts []string
	)
loop:
	for scanner.Scan() {

		cmd = scanner.Text()
		parts = strings.Split(cmd, " ")

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {
		case "stats":
			galanodes.QueryAllNodes(false)
			break
		case "help":
			printHelp()
			break

		case "exit":
			os.Exit(0)
			break loop

		case "nodes":
			galanodes.PrintNodeInfo()
			break

		case "save":
			galanodes.SaveNodeInfo()
			break

		case "discord":
			galanodes.PrintSummary(true)
			break

		case "cls":
			clearScreen()
			break

		case "reboot":
			if len(parts) == 1 {
				log.Printf("Please pass a node index or addres, all with reboot command.")
			} else {
				galanodes.RunNodeCmd(parts[1], galanodes.CmdReboot)
			}
			break

		case "find":
			if len(parts) == 1 {
				log.Printf("Please pass a string with find command.")
			} else {
				galanodes.FindNode(parts[1])
			}
			break

		case "restart":
			if len(parts) == 1 {
				log.Printf("Please pass a node index or addres with restart command.")
			} else {
				galanodes.RunNodeCmd(parts[1], galanodes.CmdRestartGalaNode)
			}
			break

		default:
			// find nodes as a default
			if len(parts) == 1 {
				galanodes.PrintNodeInfo()
			} else {
				fmt.Println("")
			}
		}

		fmt.Print("> ")
	}
}
