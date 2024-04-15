package cmd

import (
	"sgit/internal/logging"
	"testing"

	"github.com/golang/mock/gomock"
)

//go:generate mockgen -source=./cmd.go -destination=./mock_cmd_test.go -package=cmd
func TestRemoteToLocalRefresh(t *testing.T) {
	name := "repo"
	sshUrl := "repo.com"
	repo := GithubRepository{FullName: name, SshUrl: sshUrl}
	baseDir := "/Users/me"
	langBaseDir := "/Users/me/python"
	fullPath := "/Users/me/python/repo"
	python := "python"
	hash := "aa218f56b14c9653891f9e74264a383fa43fefbd"
	branch := "main"

	type testCase struct {
		name          string
		defineExpects func()
	}
	var github *MockGithub
	var git *MockGit
	var filesystem *MockFilesystem
	var tui *MockTUI

	testCases := []testCase{
		{
			name: "nonexistent local repo",
			defineExpects: func() {
				github.EXPECT().GetPrimaryLanguageForRepo(name).Return(python, nil)
				filesystem.EXPECT().Exists(fullPath).Return(false, nil)
				filesystem.EXPECT().CreateDirectory(fullPath).Return(nil)
				git.EXPECT().CloneRepo(repo, langBaseDir).Return(nil)
				tui.EXPECT().Handle(repo, UpToDate)
			},
		},
		{
			name: "existent up to date local repo",
			defineExpects: func() {
				github.EXPECT().GetPrimaryLanguageForRepo(name).Return(python, nil)
				filesystem.EXPECT().Exists(fullPath).Return(true, nil)
				git.EXPECT().GetBranchName(fullPath).Return(branch, nil)
				git.EXPECT().HasUncommittedChanges(fullPath).Return(false, nil)
				git.EXPECT().PullLatest(fullPath).Return(nil)
				git.EXPECT().HasMergeConflicts(fullPath).Return(false, nil)
				tui.EXPECT().Handle(repo, UpToDate)
			},
		},
		{
			name: "existent local repo with merge conflicts",
			defineExpects: func() {
				github.EXPECT().GetPrimaryLanguageForRepo(name).Return(python, nil)
				filesystem.EXPECT().Exists(fullPath).Return(true, nil)
				git.EXPECT().GetBranchName(fullPath).Return(branch, nil)
				git.EXPECT().HasUncommittedChanges(fullPath).Return(false, nil)
				git.EXPECT().PullLatest(fullPath).Return(nil)
				git.EXPECT().HasMergeConflicts(fullPath).Return(true, nil)
				tui.EXPECT().Handle(repo, HasMergeConflicts).Return(nil)
			},
		},
		{
			name: "existent local repo with uncommited changes",
			defineExpects: func() {
				github.EXPECT().GetPrimaryLanguageForRepo(name).Return(python, nil)
				filesystem.EXPECT().Exists(fullPath).Return(true, nil)
				git.EXPECT().GetBranchName(fullPath).Return(branch, nil)
				git.EXPECT().HasUncommittedChanges(fullPath).Return(true, nil)
				github.EXPECT().GetCommitHash(name, branch).Return(hash, nil)
				git.EXPECT().GetCommitHashes(fullPath).Return([]string{hash}, nil)
				tui.EXPECT().Handle(repo, HasUncommittedChanges).Return(nil)
			},
		},
		{
			name: "existent local repo with uncommited changes and behind upstream",
			defineExpects: func() {
				github.EXPECT().GetPrimaryLanguageForRepo(name).Return(python, nil)
				filesystem.EXPECT().Exists(fullPath).Return(true, nil)
				git.EXPECT().GetBranchName(fullPath).Return(branch, nil)
				git.EXPECT().HasUncommittedChanges(fullPath).Return(true, nil)
				github.EXPECT().GetCommitHash(name, branch).Return(hash, nil)
				git.EXPECT().GetCommitHashes(fullPath).Return([]string{"gibberish"}, nil)
				tui.EXPECT().Handle(repo, HasUncommittedChangesAndBehindUpstream).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			l := logging.New()
			github = NewMockGithub(ctrl)
			git = NewMockGit(ctrl)
			filesystem = NewMockFilesystem(ctrl)
			tui = NewMockTUI(ctrl)

			cmd := NewRefreshCommand(l, github, git, filesystem, tui, baseDir)
			tc.defineExpects()
			if err := cmd.remoteToLocalRepoRefresh(repo); err != nil {
				t.FailNow()
			}
		})
	}
}
