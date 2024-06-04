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
	Providers map[string]Provider
}

type OIDCIssuer struct {
	// The expected issuer of an OIDC token
	IssuerURL string `yaml:"issuer-url,omitempty"`
	// The expected client ID of the OIDC token
	ClientID string `yaml:"client-id"`
	// Used to determine the subject of the certificate and if additional
	// certificate values are needed
	Type string `yaml:"type"`
	// Optional, if the issuer is in a different claim in the OIDC token
	IssuerClaim string `yaml:"issuer-claim,omitempty"`
	// The domain that must be present in the subject for 'uri' issuer types
	// Also used to create an email for 'username' issuer types
	SubjectDomain string `yaml:"subject-domain,omitempty"`
	// SPIFFETrustDomain specifies the trust domain that 'spiffe' issuer types
	// issue ID tokens for. Tokens with a different trust domain will be
	// rejected.
	SPIFFETrustDomain string `yaml:"spiffe-trust-domain,omitempty"`
	// Optional, the challenge claim expected for the issuer
	// Set if using a custom issuer
	ChallengeClaim string `yaml:"challenge-claim,omitempty"`
}

type Provider struct {
	Extensions  Extensions
	Uris        []string
	Defaults    map[string]string
	OIDCIssuers []OIDCIssuer `yaml:"oidc-issuers,omitempty"`
	MetaIssuers []OIDCIssuer `yaml:"meta-issuers,omitempty"`
}

func ApplyTemplate(path string, data map[string]string, defaultData map[string]string) string {

	// Here we merge the data from was claimed by the id token with the
	// default data provided by the yaml file.
	// The order here matter because we want to override the default data
	// with the claimed data.
	mergedData := make(map[string]string)
	for k, v := range defaultData {
		mergedData[k] = v
	}
	for k, v := range data {
		mergedData[k] = v
	}

	// It checks it is a path or a raw field by
	// checking exists template syntax into the string
	if strings.Contains(path, "{{.") {
		var doc bytes.Buffer
		t := template.New("")
		p, err := t.Parse(path)
		if err != nil {
			panic(err)
		}
		err = p.Execute(&doc, mergedData)
		if err != nil {
			fmt.Println(err)
		}
		return doc.String()
	} else {
		return mergedData[path]
	}
}

func main() {
	var obj RootYaml

	yamlFile, err := os.ReadFile("pkg/providers.yaml")
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
		"account_name":       "accountname1",
		"pipeline_name":      "pipelinename1",
		"account_id":         "111222",
		"job_workflow_ref":   "refrefref",
		"ci_config_ref_uri":  "refciconfigref",
	}

	finalObj := RootYaml{Providers: make(map[string]Provider)}
	for k, provider := range obj.Providers {
		e := provider.Extensions
		d := provider.Defaults
		finalExtensions := Extensions{
			Issuer:                              ApplyTemplate(e.Issuer, runData, d),
			GithubWorkflowTrigger:               ApplyTemplate(e.GithubWorkflowTrigger, runData, d),
			GithubWorkflowSHA:                   ApplyTemplate(e.GithubWorkflowSHA, runData, d),
			GithubWorkflowName:                  ApplyTemplate(e.GithubWorkflowName, runData, d),
			GithubWorkflowRepository:            ApplyTemplate(e.GithubWorkflowRepository, runData, d),
			GithubWorkflowRef:                   ApplyTemplate(e.GithubWorkflowRef, runData, d),
			BuildSignerURI:                      ApplyTemplate(e.BuildSignerURI, runData, d),
			BuildConfigDigest:                   ApplyTemplate(e.BuildConfigDigest, runData, d),
			RunnerEnvironment:                   ApplyTemplate(e.RunnerEnvironment, runData, d),
			SourceRepositoryURI:                 ApplyTemplate(e.SourceRepositoryURI, runData, d),
			SourceRepositoryDigest:              ApplyTemplate(e.SourceRepositoryDigest, runData, d),
			SourceRepositoryRef:                 ApplyTemplate(e.SourceRepositoryRef, runData, d),
			SourceRepositoryIdentifier:          ApplyTemplate(e.SourceRepositoryIdentifier, runData, d),
			SourceRepositoryOwnerURI:            ApplyTemplate(e.SourceRepositoryOwnerURI, runData, d),
			SourceRepositoryOwnerIdentifier:     ApplyTemplate(e.SourceRepositoryOwnerIdentifier, runData, d),
			BuildConfigURI:                      ApplyTemplate(e.BuildConfigURI, runData, d),
			BuildSignerDigest:                   ApplyTemplate(e.BuildSignerDigest, runData, d),
			BuildTrigger:                        ApplyTemplate(e.BuildTrigger, runData, d),
			RunInvocationURI:                    ApplyTemplate(e.RunInvocationURI, runData, d),
			SourceRepositoryVisibilityAtSigning: ApplyTemplate(e.SourceRepositoryVisibilityAtSigning, runData, d),
		}
		finalUris := make([]string, len(provider.Uris)-1)
		for _, val := range provider.Uris {
			finalUris = append(finalUris, ApplyTemplate(val, runData, d))
		}
		provider := Provider{
			Extensions:  finalExtensions,
			Uris:        finalUris,
			OIDCIssuers: provider.OIDCIssuers,
			MetaIssuers: provider.MetaIssuers,
		}
		finalObj.Providers[k] = provider
	}
	var prettyJSON bytes.Buffer
	inrec, _ := json.Marshal(finalObj)
	json.Indent(&prettyJSON, inrec, "", "\t")
	log.Println(prettyJSON.String())
}
