package common

import (
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestGitLikeProjectSource_GetDefaultSource(t *testing.T) {

	tests := []struct {
		name                 string
		gitLikeProjectSource v1.GitLikeProjectSource
		want1                string
		want2                string
		want3                string
		wantErr              bool
	}{
		{
			name: "only one remote",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin": "url",
				},
			},
			want1:   "origin",
			want2:   "url",
			want3:   "",
			wantErr: false,
		},
		{
			name: "multiple remotes, checkoutFrom with only branch",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin": "urlO",
				},
				CheckoutFrom: &v1.CheckoutFrom{Revision: "dev"},
			},
			want1:   "origin",
			want2:   "urlO",
			want3:   "dev",
			wantErr: false,
		},
		{
			name: "multiple remotes, checkoutFrom without revision",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "upstream"},
			},
			want1:   "upstream",
			want2:   "urlU",
			want3:   "",
			wantErr: false,
		},
		{
			name: "multiple remotes, checkoutFrom with revision",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "upstream", Revision: "v1"},
			},
			want1:   "upstream",
			want2:   "urlU",
			want3:   "v1",
			wantErr: false,
		},
		{
			name: "multiple remotes, checkoutFrom with unknown remote",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "non"},
			},
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: true,
		},
		{
			name: "multiple remotes, no checkoutFrom",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
			},
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got1, got2, got3, err := GetDefaultSource(tt.gitLikeProjectSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitLikeProjectSource.GetDefaultSource() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if got1 != tt.want1 {
					t.Errorf("GitLikeProjectSource.GetDefaultSource() got1 = %v, want %v", got1, tt.want1)
				}
				if got2 != tt.want2 {
					t.Errorf("GitLikeProjectSource.GetDefaultSource() got2 = %v, want %v", got2, tt.want2)
				}
				if got3 != tt.want3 {
					t.Errorf("GitLikeProjectSource.GetDefaultSource() got2 = %v, want %v", got3, tt.want3)
				}
			}
		})
	}
}

func TestGetProjectSrcType(t *testing.T) {

	tests := []struct {
		name           string
		projectSrc     v1.ProjectSource
		wantErr        bool
		projectSrcType v1.ProjectSourceType
	}{
		{
			name: "Git project",
			projectSrc: v1.ProjectSource{
				Git: &v1.GitProjectSource{},
			},
			projectSrcType: v1.GitProjectSourceType,
			wantErr:        false,
		},
		{
			name: "Github project",
			projectSrc: v1.ProjectSource{
				Github: &v1.GithubProjectSource{},
			},
			projectSrcType: v1.GitHubProjectSourceType,
			wantErr:        false,
		},
		{
			name: "Zip project",
			projectSrc: v1.ProjectSource{
				Zip: &v1.ZipProjectSource{},
			},
			projectSrcType: v1.ZipProjectSourceType,
			wantErr:        false,
		},
		{
			name: "Custom project",
			projectSrc: v1.ProjectSource{
				Custom: &v1.CustomProjectSource{},
			},
			projectSrcType: v1.CustomProjectSourceType,
			wantErr:        false,
		},
		{
			name:       "Unknown project",
			projectSrc: v1.ProjectSource{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProjectSourceType(tt.projectSrc)
			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetProjectSrcType() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && got != tt.projectSrcType {
				t.Errorf("TestGetProjectSrcType error: project src type mismatch, expected: %v got: %v", tt.projectSrcType, got)
			}
		})
	}

}
