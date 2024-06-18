// Copyright 2024 The Sigstore Authors.
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

package ciprovider

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sigstore/fulcio/pkg/certificate"
	"github.com/sigstore/fulcio/pkg/config"
	"github.com/sigstore/fulcio/pkg/identity"
)

func claimsToString(claims map[string]interface{}) map[string]string {
	stringClaims := make(map[string]string)
	for k, v := range claims {
		stringClaims[k] = v.(string)
	}
	return stringClaims
}

func applyTemplate(path string, data map[string]string, defaultData map[string]string) string {

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

	if strings.Contains(path, "{{") {
		var doc bytes.Buffer
		t := template.New("")
		p, err := t.Parse(path)
		if err != nil {
			panic(err)
		}
		err = p.Execute(&doc, mergedData)
		if err != nil {
			panic(err)
		}
		return doc.String()
	}
	return mergedData[path]
}

type CiProvider struct {
	config.OIDCIssuer
	token *oidc.IDToken
}

func WorkflowPrincipalFromIDToken(ctx context.Context, token *oidc.IDToken) (identity.Principal, error) {
	cfg, ok := config.FromContext(ctx).GetIssuer(token.Issuer)
	if !ok {
		return nil, fmt.Errorf("configuration can not be loaded for issuer %v", token.Issuer)
	}
	return CiProvider{
		cfg,
		token,
	}, nil
}

func (p CiProvider) Name(_ context.Context) string {
	return p.token.Subject
}

func (p CiProvider) Embed(_ context.Context, cert *x509.Certificate) error {

	var claims map[string]interface{}
	if err := p.token.Claims(&claims); err != nil {
		return nil, err
	}

	e := p.CIProviderClaimsMapping
	defaults := p.Defaults
	claims := claimsToString(p.Claims)
	uris := make([]*url.URL, len(p.Uris))
	for _, value := range p.Uris {
		url, err := url.Parse(applyTemplate(value, claims, defaults))
		if err != nil {
			panic(err)
		}
		uris = append(uris, url)
	}
	// Set workflow ref URL to SubjectAlternativeName on certificate
	cert.URIs = uris

	var err error
	// Embed additional information into custom extensions
	cert.ExtraExtensions, err = certificate.Extensions{
		Issuer:                              applyTemplate(e.Issuer, claims, defaults),
		GithubWorkflowTrigger:               applyTemplate(e.GithubWorkflowTrigger, claims, defaults),
		GithubWorkflowSHA:                   applyTemplate(e.GithubWorkflowSHA, claims, defaults),
		GithubWorkflowName:                  applyTemplate(e.GithubWorkflowName, claims, defaults),
		GithubWorkflowRepository:            applyTemplate(e.GithubWorkflowRepository, claims, defaults),
		GithubWorkflowRef:                   applyTemplate(e.GithubWorkflowRef, claims, defaults),
		BuildSignerURI:                      applyTemplate(e.BuildSignerURI, claims, defaults),
		BuildConfigDigest:                   applyTemplate(e.BuildConfigDigest, claims, defaults),
		RunnerEnvironment:                   applyTemplate(e.RunnerEnvironment, claims, defaults),
		SourceRepositoryURI:                 applyTemplate(e.SourceRepositoryURI, claims, defaults),
		SourceRepositoryDigest:              applyTemplate(e.SourceRepositoryDigest, claims, defaults),
		SourceRepositoryRef:                 applyTemplate(e.SourceRepositoryRef, claims, defaults),
		SourceRepositoryIdentifier:          applyTemplate(e.SourceRepositoryIdentifier, claims, defaults),
		SourceRepositoryOwnerURI:            applyTemplate(e.SourceRepositoryOwnerURI, claims, defaults),
		SourceRepositoryOwnerIdentifier:     applyTemplate(e.SourceRepositoryOwnerIdentifier, claims, defaults),
		BuildConfigURI:                      applyTemplate(e.BuildConfigURI, claims, defaults),
		BuildSignerDigest:                   applyTemplate(e.BuildSignerDigest, claims, defaults),
		BuildTrigger:                        applyTemplate(e.BuildTrigger, claims, defaults),
		RunInvocationURI:                    applyTemplate(e.RunInvocationURI, claims, defaults),
		SourceRepositoryVisibilityAtSigning: applyTemplate(e.SourceRepositoryVisibilityAtSigning, claims, defaults),
	}.Render()
	if err != nil {
		return err
	}

	return nil
}
