package cmd

import (
	"errors"
	"sgit/internal/types"
	"testing"

	"github.com/golang/mock/gomock"
)

//go:generate mockgen -source=./cmd.go -destination=./mock_cmd_test.go -package=cmd

var (
	FULL_NAME     = "repo"
	SSH_URL       = "repo.com"
	ERROR         = errors.New("failure")
	MOCK_REPO     = types.GithubRepository{FullName: FULL_NAME, SshUrl: SSH_URL}
	BASE_DIR      = "/Users/me"
	LANG_BASE_DIR = "/Users/me/python"
	FULL_PATH     = "/Users/me/python/repo"
	PYTHON        = "python"
)

func TestRefresh(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithub := NewMockGithub(ctrl)
	mockGit := NewMockGit(ctrl)
	mockFilesystem := NewMockFilesystem(ctrl)
	cmd := newRefreshCommand(mockGithub, mockGit, mockFilesystem, BASE_DIR)

	type testCase struct {
		name          string
		defineExpects func()
		assertState   func(*repositoryState) bool
	}

	for _, tc := range []testCase{
		{
			name: "Forwards error from failed GetPrimaryLangaugeForRepo call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return("", ERROR)
			},
		},
		{
			name: "Forwards error from failed directory exists call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			name: "Forwards error from failed create directory call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mockFilesystem.EXPECT().CreateDirectory(FULL_PATH).Return(ERROR)
			},
		},
		{
			name: "Forwards error from failed clone repo call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mockFilesystem.EXPECT().CreateDirectory(FULL_PATH).Return(nil)
				mockGit.EXPECT().CloneRepo(MOCK_REPO, LANG_BASE_DIR).Return(ERROR)
			},
		},
		{
			name: "Clone repo happy path",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mockFilesystem.EXPECT().CreateDirectory(FULL_PATH).Return(nil)
				mockGit.EXPECT().CloneRepo(MOCK_REPO, LANG_BASE_DIR).Return(nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && !rs.hasLocalChanges
			},
		},
		{
			name: "Forwards error from failed HasLocalChanges call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mockGit.EXPECT().HasLocalChanges(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			name: "Forwards error from failed HasLocalChanges call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mockGit.EXPECT().HasLocalChanges(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			name: "Repo with local changes happy path",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mockGit.EXPECT().HasLocalChanges(FULL_PATH).Return(true, nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && rs.hasLocalChanges
			},
		},
		{
			name: "Forwards error from PullLatestChanges call",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mockGit.EXPECT().HasLocalChanges(FULL_PATH).Return(false, nil)
				mockGit.EXPECT().PullLatestChanges(FULL_PATH).Return(ERROR)
			},
		},
		{
			name: "Repo without local changes happy path",
			defineExpects: func() {
				mockGithub.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mockFilesystem.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mockGit.EXPECT().HasLocalChanges(FULL_PATH).Return(false, nil)
				mockGit.EXPECT().PullLatestChanges(FULL_PATH).Return(nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && !rs.hasLocalChanges
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.defineExpects()
			state, err := cmd.refresh(MOCK_REPO)

			if err != nil && err.Error() != "failure" {
				t.Fail()
				return
			}

			if tc.assertState != nil && !tc.assertState(state) {
				t.Fail()
			}

			ctrl.Finish()
		})
	}
}
