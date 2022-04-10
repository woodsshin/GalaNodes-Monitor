// Copyright 2020-present woodsshin. All rights reserved.
// Use of this source code is governed by GNU General Public License v2.0.

package galanodes

// Summary ...
type Summary struct {
	LicenseCount   int    `json:"licenseCount"`
	NodesOnline    int    `json:"nodesOnline"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	MachineID      string `json:"machineID"`
	ErrorCount     int
	State          string
	Desc           string
}

type Status struct {
	Name                 string `json:"name"`
	MyWorkloadsOnline    int    `json:"myWorkloadsOnline"`
	TotalWorkloadsOnline int    `json:"totalWorkloadsOnline"`
	MsActiveToday        int    `json:"msActiveToday"`
	LicenseCount         int    `json:"licenseCount"`
	State                string
	ErrorCount           int
}

type NodeStatus struct {
	Summary      Summary  `json:"summary"`
	Workloads    []Status `json:"workloads"`
	WorkloadsMap map[string]Status
}

const (
	ConnectionOnline            string = "Online"
	ConnectionOffline           string = "Offline"
	ConnectionRebooted          string = "Rebooted"
	ConnectionFailedAuth        string = "FailedAuth"
	ConnectionFailedCmd         string = "FailedCmd"
	ConnectionFailedJsonParsing string = "FailedJsonParsing"
)

const (
	NodeTypeFounders string = "founders"
	NodeTypeTownStar string = "townstar"
)

const (
	NodeStateOnline           string = "Online"
	NodeStateOffline          string = "Offline"
	NodeStateInActive         string = "InActive"
	NodeStateLessNodesRunning string = "LessNodesRunning"
)

const (
	CmdStats           string = "gala-node stats | jq"
	CmdConfigDevice    string = "gala-node config device"
	CmdRestartGalaNode string = "sudo systemctl restart gala-node"
	CmdRestartService  string = "sudo systemctl restart gala-node.service"
	CmdDaemon          string = "gala-node daemon"
	CmdReboot          string = "sudo reboot"
)
