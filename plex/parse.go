package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/web3-storage/go-w3s-client"
)

type Instruction struct {
	App       string            `json:"app"`
	InputCIDs []string          `json:"input_cids"`
	Container string            `json:"container"`
	Params    map[string]string `json:"params"`
	Cmd       string            `json:"cmd"`
}

func CreateInstruction(app string, instuctionFilePath, inputDirPath string, paramOverrides map[string]string) (Instruction, error) {
	instruction, err := readInstructions(app, instuctionFilePath)
	if err != nil {
		return instruction, err
	}
	instruction.Params = overwriteParams(instruction.Params, paramOverrides)
	instruction.Cmd = formatCmd(instruction.Cmd, instruction.Params)
	cid, err := createInputCID(inputDirPath, instruction.Cmd)
	if err != nil {
		return instruction, err
	}
	instruction.InputCIDs = append(instruction.InputCIDs, cid)
	return instruction, nil
}

func readInstructions(app string, filepath string) (Instruction, error) {
	fileContents, err := ioutil.ReadFile(filepath)
	var instruction Instruction
	if err != nil {
		return instruction, err
	}
	lines := strings.Split(string(fileContents), "\n")
	for _, line := range lines {
		err := json.Unmarshal([]byte(line), &instruction)
		if err != nil {
			return instruction, err
		}

		if instruction.App == app {
			return instruction, nil
		}
	}

	return instruction, fmt.Errorf("No instruction found for app %s", app)
}

func overwriteParams(defaultParams, overrideParams map[string]string) (finalParams map[string]string) {
	finalParams = make(map[string]string)
	for key, defaultVal := range defaultParams {
		if overrideVal, ok := overrideParams[key]; ok {
			finalParams[key] = overrideVal
		} else {
			finalParams[key] = defaultVal
		}
	}
	return
}

func formatCmd(cmd string, params map[string]string) (formatted string) {
	// this requires string inputs to have `%{paramKeyX}s %{paramKeyY}s"` formatting
	formatted = cmd
	for key, val := range params {
		formatted = strings.Replace(formatted, "%{"+key+"}s", fmt.Sprintf("%s", val), -1)
	}
	return
}

func createInputCID(inputDirPath string, cmd string) (string, error) {
	client, err := w3s.NewClient(
		w3s.WithEndpoint("https://api.web3.storage"),
		w3s.WithToken(os.Getenv("WEB3STORAGE_TOKEN")),
	)
	if err != nil {
		return "", err
	}
	inputDir, err := os.Open(inputDirPath)
	if err != nil {
		return "", err
	}
	cid, err := putFile(client, inputDir)
	if err != nil {
		return cid.String(), err
	}
	return cid.String(), nil
}

func createHelperFile(dirPath string, contents string) error {
	fileName := filepath.Join(dirPath, "helper.sh")
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("#!/bin/bash\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}

	err = os.Chmod(fileName, 0755)
	if err != nil {
		return err
	}

	return nil
}