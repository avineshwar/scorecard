// Copyright 2021 Security Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v2/clients"
	sce "github.com/ossf/scorecard/v2/errors"
)

const organizationsToAnalyze = 30

type contributorsData struct {
	Nodes []struct {
		User struct {
			Id            githubv4.String
			Company       githubv4.String
			Organizations struct {
				Nodes []struct {
					Login githubv4.String
				}
			} `graphql:"organizations(first:$organizationsToAnalyze)"`
		} `graphql:"... on User"`
	} `graphql:"nodes(ids:$ids)"`
}

type ContributorsHandler struct {
	GhClient     *github.Client
	GraphClient  *githubv4.Client
	data         *contributorsData
	contributors []clients.Contributor
}

func (handler *ContributorsHandler) Init(ctx context.Context, owner, repo string) error {
	contribs, _, err := handler.GhClient.Repositories.ListContributors(ctx, owner, repo, &github.ListContributorsOptions)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var nodeMap map[string]struct {
		NumContributions int
	}
	var nodeIds []githubv4.String
	for _, contrib := range contribs {
		nodeIds = append(nodeIds, githubv4.String(contrib.GetNodeId()))
		nodeMap[contrib.GetNodeId()] = contrib.GetContributions()
	}
	vars := map[string]interface{}{
		"ids":                    nodeIds,
		"organizationsToAnalyze": githubv4.Int(organizationsToAnalyze),
	}
	handler.data = new(contributorsData)
	if err := handler.GraphClient.Query(ctx, handler.data, vars); err != nil {
		// nolint: wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
	}
	// handler.contributors = contributorsFrom(nodeMap, handler.data)
	return nil
}

func (handler *ContributorsHandler) GetContributors() ([]clients.Contributor, error) {
	return handler.contributors, nil
}
