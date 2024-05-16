package packer_test

func (ts *PackerTestSuite) TestInstallPluginWithMetadata() {
	tempPluginDir, cleanup := ts.MakePluginDir("1.0.0+metadata")
	defer cleanup()

	ts.Run("metadata plugin installed must not have metadata in its path", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("plugins", "installed").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+]", grepStdout))
	})

	ts.Run("plugin with metadata should work with validate", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("validate", "./templates/simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})

	ts.Run("plugin with metadata should work with build", func() {
		ts.PackerCommand().UsePluginDir(tempPluginDir).
			SetArgs("build", "./templates/simple.pkr.hcl").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.0[^+][^\\n]+plugin:", grepStderr))
	})
}

func (ts *PackerTestSuite) TestInstallPluginPrerelease() {
	pluginPath := BuildSimplePlugin("1.0.1-alpha1", ts.T())

	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try install plugin with alpha1 prerelease - should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginDir).
			SetArgs("plugins", "install", "--path", pluginPath, "github.com/hashicorp/tester").
			Assert(MustFail(), Grep("Packer can only install plugin releases with this command", grepStdout))
	})
}

func (ts *PackerTestSuite) TestRemoteInstallWithPluginsInstall() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("install latest version of a remote plugin with packer plugins install", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups").
			Assert(MustSucceed())
	})

	ts.Run("install dev version of a remote plugin with packer plugins install - must fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "github.com/hashicorp/hashicups", "v1.0.2-dev").
			Assert(MustFail(), Grep("Remote installation of pre-release plugin versions is unsupported.", grepStdout))
	})
}

func (ts *PackerTestSuite) TestInitWithNonGithubSource() {
	pluginPath, cleanup := ts.MakePluginDir()
	defer cleanup()

	ts.Run("try installing from a non-github source, should fail", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/non_gh.pkr.hcl").
			Assert(MustFail(), Grep(`doesn't appear to be a valid "github.com" source address`, grepStdout))
	})

	ts.Run("manually install plugin to the expected source", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("plugins", "install", "--path", BuildSimplePlugin("1.0.10", ts.T()), "hubgit.com/hashicorp/tester").
			Assert(MustSucceed(), Grep("packer-plugin-tester_v1.0.10", grepStdout))
	})

	ts.Run("re-run packer init on same template, should succeed silently", func() {
		ts.PackerCommand().UsePluginDir(pluginPath).
			SetArgs("init", "./templates/non_gh.pkr.hcl").
			Assert(MustSucceed(),
				MkPipeCheck("no output in stdout").SetTester(ExpectEmptyInput()).SetStream(OnlyStdout))
	})
}