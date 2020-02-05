package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	COMMITTEES_FILE = "committees.json"
)

var CommitteesIpPortSlice []string

type Committees struct {
	Committees []Committee `json:"committees"`
}

type Committee struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

func LoadCommitteesIpPort() (ipPortSlice []string) {
	committeesFile, err := os.Open(COMMITTEES_FILE)
	defer committeesFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	var committees Committees

	byteValue, _ := ioutil.ReadAll(committeesFile)

	json.Unmarshal(byteValue, &committees)


	for i := 0; i < len(committees.Committees); i++ {
		ipPortSlice = append(ipPortSlice, committees.Committees[i].Ip + ":" +  committees.Committees[i].Port)
	}



	return ipPortSlice
}
