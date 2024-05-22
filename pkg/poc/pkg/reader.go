// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Extensions struct {
	Issuer                              string // OID 1.3.6.1.4.1.57264.1.8 and 1.3.6.1.4.1.57264.1.1 (Deprecated)
	GithubWorkflowTrigger               string `yaml:"github-workflow-trigger"`                 // OID 1.3.6.1.4.1.57264.1.2
	GithubWorkflowSHA                   string `yaml:"github-workflow-sha"`                     // OID 1.3.6.1.4.1.57264.1.3
	GithubWorkflowName                  string `yaml:"github-workflow-name"`                    // OID 1.3.6.1.4.1.57264.1.4
	GithubWorkflowRepository            string `yaml:"github-workflow-repository"`              // OID 1.3.6.1.4.1.57264.1.5
	GithubWorkflowRef                   string `yaml:"github-workflow-ref"`                     // 1.3.6.1.4.1.57264.1.6
	BuildSignerURI                      string `yaml:"build-signer-uri"`                        // 1.3.6.1.4.1.57264.1.9
	BuildSignerDigest                   string `yaml:"build-signer-digest"`                     // 1.3.6.1.4.1.57264.1.10
	RunnerEnvironment                   string `yaml:"runner-environment"`                      // 1.3.6.1.4.1.57264.1.11
	SourceRepositoryURI                 string `yaml:"source-repository-uri"`                   // 1.3.6.1.4.1.57264.1.12
	SourceRepositoryDigest              string `yaml:"source-repository-digest"`                // 1.3.6.1.4.1.57264.1.13
	SourceRepositoryRef                 string `yaml:"source-repository-ref"`                   // 1.3.6.1.4.1.57264.1.14
	SourceRepositoryIdentifier          string `yaml:"source-repository-identifier"`            // 1.3.6.1.4.1.57264.1.15
	SourceRepositoryOwnerURI            string `yaml:"source-repository-owner-uri"`             // 1.3.6.1.4.1.57264.1.16
	SourceRepositoryOwnerIdentifier     string `yaml:"source-repository-owner-identifier"`      // 1.3.6.1.4.1.57264.1.17
	BuildConfigURI                      string `yaml:"build-config-uri"`                        // 1.3.6.1.4.1.57264.1.18
	BuildConfigDigest                   string `yaml:"build-config-digest"`                     // 1.3.6.1.4.1.57264.1.19
	BuildTrigger                        string `yaml:"build-trigger"`                           // 1.3.6.1.4.1.57264.1.20
	RunInvocationURI                    string `yaml:"run-invocation-uri"`                      // 1.3.6.1.4.1.57264.1.21
	SourceRepositoryVisibilityAtSigning string `yaml:"source-repository-visibility-at-signing"` // 1.3.6.1.4.1.57264.1.22
}

type RootYaml struct {
	Providers map[string]Provider `yaml:"providers"`
}

type Provider struct {
	extensions Extensions        `yaml:"extensions"`
	uris       []string          `yaml:"uris"`
	defaults   map[string]string `yaml:"defaults"`
}

func ApplyTemplate(path string, data map[string]string) string {

	// It checks it is a path or a raw field by
	// checking exists template syntax into the string
	if strings.Contains(path, "{{.") {
		var doc bytes.Buffer
		t := template.New("")
		p, err := t.Parse(path)
		if err != nil {
			panic(err)
		}
		err = p.Execute(&doc, data)
		if err != nil {
			fmt.Println(err)
		}
		return doc.String()
	} else {
		return data[path]
	}
}

func main() {
	var obj RootYaml

	yamlFile, err := os.ReadFile("pkg/providers.yaml")
	fmt.Printf("%v\n", yamlFile)
	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &obj)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	runData := map[string]string{
		"repository":         "ossf/scorecard",
		"run_id":             "123",
		"run_attempt":        "345",
		"job_workflow_sha":   "1a2b3c",
		"project_path":       "project/path",
		"job_id":             "1233",
		"workflow_id":        "123123",
		"pipeline_id":        "321321",
		"runner_environment": "staging",
		"sha":                "3c2b1a",
		"repository_id":      "1232321232",
		"project_id":         "333333",
		"scm_repo_url":       "scm/repo/url",
		"scm_ref":            "scmref",
	}
	fmt.Printf("%v\n", obj)
	for _, provider := range obj.Providers {
		v := provider.extensions
		fmt.Printf("%v\n", provider)
		extensions := Extensions{
			Issuer:                              ApplyTemplate(v.Issuer, runData),
			GithubWorkflowTrigger:               ApplyTemplate(v.GithubWorkflowTrigger, runData),
			GithubWorkflowSHA:                   ApplyTemplate(v.GithubWorkflowSHA, runData),
			GithubWorkflowName:                  ApplyTemplate(v.GithubWorkflowName, runData),
			GithubWorkflowRepository:            ApplyTemplate(v.GithubWorkflowRepository, runData),
			GithubWorkflowRef:                   ApplyTemplate(v.GithubWorkflowRef, runData),
			BuildSignerURI:                      ApplyTemplate(v.BuildSignerURI, runData),
			BuildConfigDigest:                   ApplyTemplate(v.BuildConfigDigest, runData),
			RunnerEnvironment:                   ApplyTemplate(v.RunnerEnvironment, runData),
			SourceRepositoryURI:                 ApplyTemplate(v.SourceRepositoryURI, runData),
			SourceRepositoryDigest:              ApplyTemplate(v.SourceRepositoryDigest, runData),
			SourceRepositoryRef:                 ApplyTemplate(v.SourceRepositoryRef, runData),
			SourceRepositoryIdentifier:          ApplyTemplate(v.SourceRepositoryIdentifier, runData),
			SourceRepositoryOwnerURI:            ApplyTemplate(v.SourceRepositoryOwnerURI, runData),
			SourceRepositoryOwnerIdentifier:     ApplyTemplate(v.SourceRepositoryOwnerIdentifier, runData),
			BuildConfigURI:                      ApplyTemplate(v.BuildConfigURI, runData),
			BuildSignerDigest:                   ApplyTemplate(v.BuildSignerDigest, runData),
			BuildTrigger:                        ApplyTemplate(v.BuildTrigger, runData),
			RunInvocationURI:                    ApplyTemplate(v.RunInvocationURI, runData),
			SourceRepositoryVisibilityAtSigning: ApplyTemplate(v.SourceRepositoryVisibilityAtSigning, runData),
		}
		var prettyJSON bytes.Buffer
		inrec, _ := json.Marshal(extensions)
		json.Indent(&prettyJSON, inrec, "", "\t")
		log.Println(prettyJSON.String())
	}
}
