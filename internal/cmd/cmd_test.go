package cmd

import (
	"errors"
	mock_interfaces "sgit/internal/intefaces/mocks"
	"sgit/internal/types"

	"testing"

	"github.com/golang/mock/gomock"
)

var (
	FULL_NAME = "repo"
	SSH_URL   = "repo.com"
	ERROR     = errors.New("failure")
	MOCK_REPO = types.GithubRepository{FullName: FULL_NAME, SshUrl: SSH_URL}
	BASE_PATH = "/Users/me"
	FULL_PATH = "/Users/me/python/repo"
	PYTHON    = "python"
)

func TestRefresh(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubClient := mock_interfaces.NewMockGithubClient(ctrl)
	mockGitClient := mock_interfaces.NewMockGitClient(ctrl)
	mockFilesystem := mock_interfaces.NewMockFilesystemClient(ctrl)
	cmd := newRefreshCommand(mockGithubClient, mockGitClient, mockFilesystem, BASE_PATH)

	type TestCase struct {
		description   string
		defineExpects func(*mock_interfaces.MockGithubClient, *mock_interfaces.MockGitClient, *mock_interfaces.MockFilesystemClient)
		assertState   func(*repositoryState) bool
	}

	for _, tc := range []TestCase{
		{
			description: "Forwards error from failed GetPrimaryLangaugeForRepo call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return("", ERROR)
			},
		},
		{
			description: "Forwards error from failed directory exists call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			description: "Forwards error from failed create directory call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mfc.EXPECT().CreateDirectory(FULL_PATH).Return(ERROR)
			},
		},
		{
			description: "Forwards error from failed clone repo call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mfc.EXPECT().CreateDirectory(FULL_PATH).Return(nil)
				mgc.EXPECT().CloneRepo(MOCK_REPO, FULL_PATH).Return(ERROR)
			},
		},
		{
			description: "Clone repo happy path",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(false, nil)
				mfc.EXPECT().CreateDirectory(FULL_PATH).Return(nil)
				mgc.EXPECT().CloneRepo(MOCK_REPO, FULL_PATH).Return(nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && !rs.hasLocalChanges
			},
		},
		{
			description: "Forwards error from failed HasLocalChanges call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mgc.EXPECT().HasLocalChanges(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			description: "Forwards error from failed HasLocalChanges call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mgc.EXPECT().HasLocalChanges(FULL_PATH).Return(false, ERROR)
			},
		},
		{
			description: "Repo with local changes happy path",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mgc.EXPECT().HasLocalChanges(FULL_PATH).Return(true, nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && rs.hasLocalChanges
			},
		},
		{
			description: "Forwards error from PullLatestChanges call",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mgc.EXPECT().HasLocalChanges(FULL_PATH).Return(false, nil)
				mgc.EXPECT().PullLatestChanges(FULL_PATH).Return(ERROR)
			},
		},
		{
			description: "Repo without local changes happy path",
			defineExpects: func(mghc *mock_interfaces.MockGithubClient, mgc *mock_interfaces.MockGitClient, mfc *mock_interfaces.MockFilesystemClient) {
				mghc.EXPECT().GetPrimaryLanguageForRepo(FULL_NAME).Return(PYTHON, nil)
				mfc.EXPECT().Exists(FULL_PATH).Return(true, nil)
				mgc.EXPECT().HasLocalChanges(FULL_PATH).Return(false, nil)
				mgc.EXPECT().PullLatestChanges(FULL_PATH).Return(nil)
			},
			assertState: func(rs *repositoryState) bool {
				return rs.path == FULL_PATH && rs.r.FullName == "repo" && rs.r.SshUrl == "repo.com" && rs.r.Name() == "repo" && !rs.hasLocalChanges
			},
		},
	} {
		tc.defineExpects(mockGithubClient, mockGitClient, mockFilesystem)
		state, err := cmd.refresh(MOCK_REPO)

		if err != nil && err.Error() != "failure" {
			t.Fail()
			return
		}

		if tc.assertState != nil && !tc.assertState(state) {
			t.Fail()
		}

		ctrl.Finish()
	}
}
