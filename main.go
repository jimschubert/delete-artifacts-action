// Copyright 2020 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	app "github.com/jimschubert/delete-artifacts"
	act "github.com/sethvargo/go-githubactions"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	log.Infof("delete-artifacts-action %s (%s)", version, commit)
	log.Infof("https://github.com/jimschubert/delete-artifacts-action")
	fmt.Println()

	githubToken := act.GetInput("GITHUB_TOKEN")
	if githubToken == "" {
		var ok bool
		// allow for local testing
		githubToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			log.Fatal("Missing input 'GITHUB_TOKEN' in action configuration.")
		}
	}

	fullRepo := act.GetInput("GITHUB_REPOSITORY")
	if !strings.Contains(fullRepo, "/") {
		log.WithFields(log.Fields{"GITHUB_REPOSITORY": fullRepo}).Fatal("Invalid GITHUB_REPOSITORY. Must be in the format: owner/repo")
	}

	runIdInput := act.GetInput("run_id")
	minBytesInput := act.GetInput("min_bytes")
	maxBytesInput := act.GetInput("max_bytes")
	nameInput := act.GetInput("artifact_name")
	patternInput := act.GetInput("pattern")
	activeDurationInput := act.GetInput("active_duration")
	dryRunInput := act.GetInput("dry_run")
	logLevel := act.GetInput("log_level")

	log.WithFields(log.Fields{
		"runIdInput":          runIdInput,
		"minBytesInput":       minBytesInput,
		"maxBytesInput":       maxBytesInput,
		"nameInput":           nameInput,
		"patternInput":        patternInput,
		"activeDurationInput": activeDurationInput,
		"dryRunInput":         dryRunInput,
		"logLevel":            logLevel,
	}).Info("Input arguments")

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.DebugLevel
	}

	_ = os.Setenv("GITHUB_TOKEN", githubToken)
	_ = os.Setenv("LOG_LEVEL", level.String())

	repoParts := strings.Split(fullRepo, "/")
	owner := &repoParts[0]
	repo := &repoParts[1]

	runId := parseOptionalInt64(runIdInput)

	minBytes, err := strconv.ParseInt(minBytesInput, 10, 32)
	if err != nil {
		log.Error("Unable to parse min_bytes")
	}

	maxBytes := parseOptionalInt64(maxBytesInput)

	dryRun, err := strconv.ParseBool(dryRunInput)
	if err != nil {
		dryRun = false
	}

	instance, err := app.New(owner, repo, runId, minBytes, maxBytes, nameInput, patternInput, activeDurationInput, dryRun)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to initialize the action.")
		return
	}

	err = instance.Run()
	if err != nil {
		log.Error("Failed to run the action.")
	}
}

func parseOptionalInt64(input string) *int64 {
	var result *int64 = nil
	if len(input) > 0 {
		r, err := strconv.ParseInt(input, 10, 64)
		if err != nil && r > 0 {
			result = &r
		} else {
			result = nil
		}
	}
	return result
}
