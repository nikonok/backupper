package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nikonok/backupper/helpers"
)

func runViewLog(appCfg *helpers.AppConfig) {
	fmt.Println("Running view logs mode. Next line is log from a log file!")

	if len(appCfg.ViewCfg.Date) != 0 || appCfg.ViewCfg.Regex != helpers.DEFAULT_REGEX {
		filterLogs(appCfg)
	} else {
		viewLogs(appCfg)
	}
}

func viewLogs(appCfg *helpers.AppConfig) {
	file, err := os.Open(appCfg.LoggerFilePath)
	if err != nil {
		fmt.Println("Error opening log file: " + err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func filterLogs(appCfg *helpers.AppConfig) {
	file, err := os.Open(appCfg.LoggerFilePath)
	if err != nil {
		fmt.Println("Error opening log file: " + err.Error())
		return
	}
	defer file.Close()

	regex, err := regexp.Compile(appCfg.ViewCfg.Regex)
	if err != nil {
		fmt.Println("Invalid regex: " + err.Error())
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, appCfg.ViewCfg.Date) && regex.MatchString(line) {
			fmt.Println(line)
		}
	}
}
