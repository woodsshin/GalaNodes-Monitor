// Copyright 2020-present woodsshin. All rights reserved.
// Use of this source code is governed by GNU General Public License v2.0.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	nodeconfig "galamonitor/config"
	"galamonitor/galanodes"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	nodeConfig nodeconfig.NodeConfig

	clear map[string]func() //create a map for storing clear funcs

	nodeIdx   int
	nodeTypes [2]string
	nodeList  []galanodes.NodeStatus
	nodeMap   map[string]galanodes.NodeStatus
	alertMap  map[string]int64

	wg            sync.WaitGroup
	lastQueryTime int64
)

func Init() {
	nodeMap = make(map[string]galanodes.NodeStatus)
	alertMap = make(map[string]int64)

	// this is to display nodes in order. new node has to be added to nodeTypes.
	nodeTypes[0] = galanodes.NodeTypeFounders
	nodeTypes[1] = galanodes.NodeTypeTownStar

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

	var (
		err error
	)

	Init()

	nodeConfig, err = nodeconfig.GetNodeConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("Error: %v", err))
	}

	go runConsole()

	go regularReport()

	for {
		queryAllNodes(true)

		// save the last query time
		lastQueryTime = time.Now().UTC().Unix()

		var duration time.Duration
		duration = time.Duration(nodeConfig.Settings.MonitorInterval)
		time.Sleep(duration * time.Second)
	}
}

func regularReport() {

	for {
		var duration time.Duration
		duration = time.Duration(nodeConfig.Settings.RegularReportInterval)
		time.Sleep(duration * time.Second)

		printSummary(true)
	}
}

func printSummary(discordReport bool) {
	var (
		onlineFoundersNodes      int
		offlineFoundersNodes     int
		monitoringFoundersNodes  int
		totalOnlineFoundersNodes int
		onlineTownNodes          int
		offlineTownNodes         int
		monitoringTownNodes      int
		totalOnlineTownNodes     int
	)

	color := 0

	// get registered nodes count
	for _, nodeInfo := range nodeConfig.Servers {
		for _, v := range nodeInfo.Nodes {
			switch v {
			case galanodes.NodeTypeFounders:
				monitoringFoundersNodes++
				break
			case galanodes.NodeTypeTownStar:
				monitoringTownNodes++
				break
			}
		}
	}

	// get active nodes count
	for _, nodeInfo := range nodeMap {
		for _, v := range nodeInfo.WorkloadsMap {

			switch v.Name {
			case galanodes.NodeTypeFounders:
				if v.State == galanodes.NodeStateOnline {
					onlineFoundersNodes++
					totalOnlineFoundersNodes = v.TotalWorkloadsOnline
				} else {
					offlineFoundersNodes++
					color = 15158332
				}
				break
			case galanodes.NodeTypeTownStar:
				if v.State == galanodes.NodeStateOnline {
					onlineTownNodes++
					totalOnlineTownNodes = v.TotalWorkloadsOnline
				} else {
					offlineTownNodes++
					color = 15158332
				}
				break
			}

		}
	}
	report := fmt.Sprintf("founders : %v/%v total online nodes : %v\ttownstar : %v/%v total online nodes : %v",
		onlineFoundersNodes, monitoringFoundersNodes, totalOnlineFoundersNodes,
		onlineTownNodes, monitoringTownNodes, totalOnlineTownNodes)

	log.Println(report)

	if discordReport {
		report = fmt.Sprintf("founders : %v/%v\\ttotal online nodes : %v\\ntownstar : %v/%v\\ttotal online nodes : %v",
			onlineFoundersNodes, monitoringFoundersNodes, totalOnlineFoundersNodes,
			onlineTownNodes, monitoringTownNodes, totalOnlineTownNodes)

		embedString := fmt.Sprintf("{\"embeds\":[{\"title\": \"%v\",\"author\":{\"name\":\"Nodes report\",\"icon_url\":\"https://app.gala.games/_nuxt/img/icon_gala_cube.a0b796d.png\"},\"color\":%v}]}",
			report, color)

		sendDiscordMessage(embedString)
	}
}

func printNodeError(idx int, nodeInfo galanodes.NodeStatus, nodesetting nodeconfig.NodeSettings, reason string) {
	if len(reason) > 0 {
		fmt.Printf("\n%v. %v(%v)%v\n", idx, nodesetting.Name, nodesetting.Address, reason)
	}

	// update node state
	preNodeInfo := nodeMap[nodesetting.Address]
	nodeInfo.Summary.ErrorCount = preNodeInfo.Summary.ErrorCount + 1

	nodeMap[nodesetting.Address] = nodeInfo

	if len(nodeConfig.Settings.WebHookUrl) == 0 {
		return
	}

	// error tolerance
	if nodeInfo.Summary.ErrorCount < nodeConfig.Settings.ErrorTolerance {
		return
	}

	// reset error count
	nodeInfo.Summary.ErrorCount = 0
	nodeMap[nodesetting.Address] = nodeInfo

	// update notify time
	nextNotifyTime := alertMap[nodesetting.Address]
	if nextNotifyTime > time.Now().UTC().Unix() {
		return
	}

	alertMap[nodesetting.Address] = time.Now().UTC().Unix() + int64(nodeConfig.Settings.DiscordNotifySnooze)

	var machineId string
	if len(nodeInfo.Summary.MachineID) > 0 {
		machineIds := strings.Split(nodeInfo.Summary.MachineID, ":")
		machineId = machineIds[0]
	}

	// generate discord message
	var nodes string
	for _, nodeType := range nodeTypes {
		v := nodeInfo.WorkloadsMap[nodeType]
		if len(v.Name) == 0 {
			continue
		}
		nodes += fmt.Sprintf("%v active : %v\\nLicenses : %v/%v\\nNode state : %v\\n\\n",
			v.Name, getTimeStamp(v.MsActiveToday), v.MyWorkloadsOnline, v.LicenseCount, v.State)
	}

	embedString := fmt.Sprintf("{\"embeds\":[{\"title\": \"Online nodes : %v\\nLicenses count : %v\\nMachine ID : %v\\n\\n%v\",\"author\":{\"name\":\"%v. %v(%v) ver %v connection state : %v\",\"icon_url\":\"https://app.gala.games/_nuxt/img/icon_gala_cube.a0b796d.png\"},\"color\":15158332}]}",
		nodeInfo.Summary.NodesOnline, nodeInfo.Summary.LicenseCount,
		machineId, nodes,
		idx, nodesetting.Name, nodesetting.Address,
		nodeInfo.Summary.CurrentVersion, nodeInfo.Summary.State)

	sendDiscordMessage(embedString)
}

func sendDiscordMessage(embedString string) {
	var jsonStr = []byte(embedString)
	req, err := http.NewRequest("POST", nodeConfig.Settings.WebHookUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func printNodeInfo() {
	idx := 1
	for _, nodesettings := range nodeConfig.Servers {
		nodeInfo := nodeMap[nodesettings.Address]
		fmt.Printf("%v. %v(%v) licenses %v/%v state %v ver %v %v\n",
			idx, nodesettings.Name, nodesettings.Address,
			nodeInfo.Summary.NodesOnline, nodeInfo.Summary.LicenseCount, nodeInfo.Summary.State,
			nodeInfo.Summary.CurrentVersion, nodeInfo.Summary.MachineID)

		for _, nodeType := range nodeTypes {
			v := nodeInfo.WorkloadsMap[nodeType]
			if len(v.Name) == 0 {
				continue
			}
			fmt.Printf("%v active : %v licenses : %v/%v state : %v\n",
				v.Name, getTimeStamp(v.MsActiveToday), v.MyWorkloadsOnline, v.LicenseCount, v.State)
		}
		fmt.Println("")
		idx++
	}
	nextQueryTime := int64(nodeConfig.Settings.MonitorInterval) - (time.Now().UTC().Unix() - lastQueryTime)
	if nextQueryTime <= 0 {
		log.Println("running nodes stats")
	} else {
		log.Printf("next nodes stats will run in %v seconds", nextQueryTime)
	}
}

func runNodeCmd(key string, cmd string) {
	var (
		nodeIdx     int
		isKeyIdx    bool
		nodesetting nodeconfig.NodeSettings
	)
	value, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		isKeyIdx = false
	} else {
		isKeyIdx = true
		// non zero based index
		nodeIdx = int(value) - 1
	}

	if strings.ToLower(key) == "all" {
		log.Printf("N/A")
	} else {
		if isKeyIdx {
			// index of nodes
			if nodeIdx >= len(nodeConfig.Servers) {
				log.Printf("failed to reboot node. not found key : %v", key)
				return
			}
			nodesetting = nodeConfig.Servers[nodeIdx]
		} else {
			// find idx
			var found bool
			for idx, nodesetting := range nodeConfig.Servers {
				if nodesetting.Address == key {
					nodeIdx = idx
					found = true
					break
				}
			}
			if found == false {
				log.Printf("failed to reboot node. not found key : %v", key)
				return
			}
		}

		queryNode(nodeIdx, cmd, nodesetting, false, false)
	}
}

func findNode(key string) {
	idx := 0
	foundNodes := 0
	for _, nodesettings := range nodeConfig.Servers {
		nodeInfo := nodeMap[nodesettings.Address]
		idx++

		// find string
		hasNode := false
		if strings.Contains(nodesettings.Name, key) ||
			strings.Contains(nodesettings.Address, key) ||
			strings.Contains(nodeInfo.Summary.MachineID, key) {
			hasNode = true
		}

		if hasNode == false {
			continue
		}

		foundNodes++

		fmt.Printf("%v. %v(%v) licenses %v/%v state %v ver %v %v\n",
			idx, nodesettings.Name, nodesettings.Address,
			nodeInfo.Summary.NodesOnline, nodeInfo.Summary.LicenseCount, nodeInfo.Summary.State,
			nodeInfo.Summary.CurrentVersion, nodeInfo.Summary.MachineID)

		for _, nodeType := range nodeTypes {
			v := nodeInfo.WorkloadsMap[nodeType]
			if len(v.Name) == 0 {
				continue
			}
			fmt.Printf("%v active : %v licenses : %v/%v state : %v\n",
				v.Name, getTimeStamp(v.MsActiveToday), v.MyWorkloadsOnline, v.LicenseCount, v.State)
		}
		fmt.Println("")
	}

	log.Printf("found %v nodes.", foundNodes)
}

func saveNodeInfo() {
	fo, err := os.Create("output.csv")
	if err != nil {
		fmt.Println(err)
		return
	}

	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
		}
	}()

	line := fmt.Sprintf("no.,name,address,FN active,TN active,state,FN online,FN licenses,TN online, TN licenses,ver,machine id\n")
	if _, err := fo.WriteString(line); err != nil {
		fmt.Println(err)
		return
	}

	idx := 1
	for _, nodesettings := range nodeConfig.Servers {
		nodeInfo := nodeMap[nodesettings.Address]
		var (
			fnactive   string
			tsactive   string
			fnlicenses string
			tslicenses string
		)

		for _, nodeType := range nodeTypes {
			v := nodeInfo.WorkloadsMap[nodeType]
			if len(v.Name) == 0 {
				continue
			}
			switch v.Name {
			case galanodes.NodeTypeFounders:
				fnactive = getTimeStamp(v.MsActiveToday)
				fnlicenses = fmt.Sprintf("%v,%v", v.MyWorkloadsOnline, v.LicenseCount)
				break
			case galanodes.NodeTypeTownStar:
				tsactive = getTimeStamp(v.MsActiveToday)
				tslicenses = fmt.Sprintf("%v,%v", v.MyWorkloadsOnline, v.LicenseCount)
				break
			}
		}

		line = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
			idx, nodesettings.Name, nodesettings.Address,
			fnactive, tsactive, nodeInfo.Summary.State, fnlicenses, tslicenses,
			nodeInfo.Summary.CurrentVersion, nodeInfo.Summary.MachineID)

		if _, err := fo.WriteString(line); err != nil {
			fmt.Println(err)
			return
		}

		idx++
	}
	log.Printf("saved output.csv")
}

func queryAllNodes(updateActiveTime bool) {
	nodeIdx = 0
	nodeList = make([]galanodes.NodeStatus, 0)

	len := len(nodeConfig.Servers)
	log.Printf("running %v nodes stats", len)

	wg.Add(len)

	for _, nodesetting := range nodeConfig.Servers {
		nodeIdx++
		go queryNode(nodeIdx, galanodes.CmdStats, nodesetting, true, updateActiveTime)
	}

	wg.Wait()

	fmt.Println("")
	// print summary
	printSummary(false)
}

func queryNode(idx int, cmd string, nodesetting nodeconfig.NodeSettings, wgCount bool, updateActiveTime bool) {
	var (
		err      error
		auth     goph.Auth
		client   *goph.Client
		timeout  time.Duration
		nodeInfo galanodes.NodeStatus
	)

	if wgCount {
		defer wg.Done()
	}

	// start progress
	if updateActiveTime == false {
		fmt.Printf(".")
	}

	if len(nodesetting.PrivateKeypath) != 0 {
		auth, err = goph.Key(nodesetting.PrivateKeypath, getPassphrase(false))
		if err != nil {
			fmt.Printf("\nfailed to authenticate(PrivateKeypath) : %v", err)
			nodeInfo.Summary.State = galanodes.ConnectionFailedAuth
			printNodeError(idx, nodeInfo, nodesetting, "failed to authenticate(PrivateKeypath)")
			return
		}
	} else {
		auth = goph.Password(nodesetting.Password)
	}

	client, err = goph.NewConn(&goph.Config{
		User:     nodesetting.Username,
		Addr:     nodesetting.Address,
		Port:     nodesetting.Port,
		Auth:     auth,
		Callback: VerifyHost,
	})

	if err != nil {
		fmt.Printf("\nfailed to connect SSH : %v", err)
		nodeInfo.Summary.State = galanodes.ConnectionOffline
		nodeInfo.Summary.Desc = "failed to connect SSH"
		printNodeError(idx, nodeInfo, nodesetting, "failed to connect SSH")
		return
	}

	// Close client net connection
	defer client.Close()

	ctx := context.Background()
	// create a context with timeout, if supplied in the argumetns
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	out, err := client.RunContext(ctx, cmd)

	if err != nil {
		fmt.Printf("\nfailed to run %v : %v", cmd, err)
		nodeInfo.Summary.State = galanodes.ConnectionFailedCmd
		nodeInfo.Summary.Desc = fmt.Sprintf("failed to run %v", cmd)
		printNodeError(idx, nodeInfo, nodesetting, nodeInfo.Summary.Desc)
		return
	}

	if cmd == galanodes.CmdStats {
		err = json.Unmarshal(out, &nodeInfo)
		if err != nil {
			fmt.Printf("\nfailed to parse stats json : %v", err)
			nodeInfo.Summary.State = galanodes.ConnectionFailedJsonParsing
			nodeInfo.Summary.Desc = string(out)
			printNodeError(idx, nodeInfo, nodesetting, string(out))
			return
		}
		parseNodeStats(idx, nodeInfo, nodesetting, updateActiveTime)
	}

	// end progress
	fmt.Printf(".")

	if cmd == galanodes.CmdRestartGalaNode {
		fmt.Println("")
		log.Printf("restarted %v(%v)", nodesetting.Name, nodesetting.Address)
	} else if cmd == galanodes.CmdReboot {
		fmt.Println("")
		log.Printf("rebooted %v(%v)", nodesetting.Name, nodesetting.Address)
		preNodeInfo := nodeMap[nodesetting.Address]
		preNodeInfo.Summary.State = galanodes.ConnectionRebooted
		nodeMap[nodesetting.Address] = preNodeInfo
	}
}

func hasNodes(nodeType string, nodesetting nodeconfig.NodeSettings) bool {
	for _, v := range nodesetting.Nodes {
		if v == nodeType {
			return true
		}
	}
	return false
}

func parseNodeStats(idx int, nodeInfo galanodes.NodeStatus, nodesetting nodeconfig.NodeSettings, updateActiveTime bool) {

	nodeInfo.Summary.State = galanodes.ConnectionOnline

	preNodeInfo := nodeMap[nodesetting.Address]

	// update nodes state
	hasError := false
	for _, v := range nodeInfo.Workloads {
		// check tracking nodes are running on this server
		if !hasNodes(v.Name, nodesetting) {
			continue
		}

		if nodeInfo.WorkloadsMap == nil {
			nodeInfo.WorkloadsMap = make(map[string]galanodes.Status)
		}

		v.State = galanodes.NodeStateOnline
		if len(preNodeInfo.Summary.MachineID) > 0 {

			// check active time changed
			status := preNodeInfo.WorkloadsMap[v.Name]
			//log.Printf("%v(%v) %v old %v new %v", nodesetting.Name, nodesetting.Address,
			//	v.Name, status.MsActiveToday, v.MsActiveToday)
			if updateActiveTime {
				if status.MsActiveToday == v.MsActiveToday {
					v.State = galanodes.NodeStateInActive
				}
			} else {
				// not update active time. run by command
				v.MsActiveToday = status.MsActiveToday
			}

			// running nodes less than license count
			if v.MyWorkloadsOnline < v.LicenseCount {
				if nodeInfo.Summary.State == galanodes.NodeStateOnline {
					v.State = galanodes.NodeStateLessNodesRunning
				}
			}
			if v.State != galanodes.NodeStateOnline {
				// check errortolerance
				v.ErrorCount = status.ErrorCount + 1
				if v.ErrorCount >= nodeConfig.Settings.ErrorTolerance {
					hasError = true
					v.ErrorCount = 0
				}
			}
		}

		nodeInfo.WorkloadsMap[v.Name] = v
		//log.Printf("%v. %v(%v) %v %v", k, nodesetting.Name, nodesetting.Address, v.Name, getTimeStamp(v.MsActiveToday))
	}

	// cache node information
	nodeList = append(nodeList, nodeInfo)
	nodeMap[nodesetting.Address] = nodeInfo

	if hasError {
		printNodeError(idx, nodeInfo, nodesetting, "")
	}
}

func getTimeStamp(activeTime int) string {
	sec := activeTime / 1000
	timeStamp := fmt.Sprintf("%02dh %02dm %02ds", sec/60/60, sec/60%60, sec%60)
	return timeStamp
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	// bypass for now
	return nil
}

func clearScreen() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func askPass(msg string) string {

	fmt.Print(msg)

	pass, err := terminal.ReadPassword(0)

	if err != nil {
		panic(err)
	}

	fmt.Println("")

	return strings.TrimSpace(string(pass))
}

func getPassphrase(ask bool) string {

	if ask {

		return askPass("Enter Private Key Passphrase: ")
	}

	return ""
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
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
			queryAllNodes(false)
			break
		case "help":
			printHelp()
			break

		case "exit":
			os.Exit(0)
			break loop

		case "nodes":
			printNodeInfo()
			break

		case "save":
			saveNodeInfo()
			break

		case "discord":
			printSummary(true)
			break

		case "cls":
			clearScreen()
			break

		case "reboot":
			if len(parts) == 1 {
				log.Printf("Please pass a node index or addres, all with reboot command.")
			} else {
				runNodeCmd(parts[1], galanodes.CmdReboot)
			}
			break

		case "find":
			if len(parts) == 1 {
				log.Printf("Please pass a string with find command.")
			} else {
				findNode(parts[1])
			}
			break

		case "restart":
			if len(parts) == 1 {
				log.Printf("Please pass a node index or addres with restart command.")
			} else {
				runNodeCmd(parts[1], galanodes.CmdRestartGalaNode)
			}
			break

		default:
			// find nodes as a default
			if len(parts) == 1 {
				printNodeInfo()
			} else {
				fmt.Println("")
			}
		}

		fmt.Print("> ")
	}
}
