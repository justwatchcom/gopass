## Releases

Note: Only members who have at least `write` [access](https://github.com/gopasspw/gopass/settings/access) to the gopass repo can create releases.

Gopass uses [goreleaser](https://goreleaser.com/) to create releases. The configuration is in the file [`.goreleaser.yml`](../.goreleaser.yml).

Goreleaser automates most but not all steps of a new release.

Note: We use semantic versioning for the command line interface and tool behaviour
but not for the API (i.e. `pkg/gopass`). Maintaining both properties in the
same repository / Go module is too cumbersome.

### Development overview

Preparing and creating a new release requires a number of steps.
Starting right after the previous release these are roughly:

* Create a new Milestone in the [GitHub issue tracker](https://github.com/gopasspw/gopass/milestones)
* Create or assign issues for the next Milestone
  * We use [Slack](https://gopassworkspace.slack.com/) to discuss prioritization and responsibilities
* After enough changes have been accumulated on the master branch we might agree to cut a new release
* Now we survey open issues for any "blockers" that should make it into the next release
  * This usually either happens in Slack or on semi-regular video calls
* After all blockers have been addressed we move the remaining issues to the next milestone and prepare the release

### Cutting a release

Once master is in good shape we need to update some metadata, build and push the release.

* Determine the next release version
  * Usually we bump the patch component
  * If we're shipping possibly disruptive changes we bump the minor component
  * So far we've never bumped the major component
* Grep all `RELEASE_NOTES` from the commit messages and prepend a new section to the [CHANGELOG](../CHANGELOG.md)
  * Major changes should be detailed in a few sentences
* Write the new version to the VERSION file
* Write the new version to the version.go file
* Commit these changes to a branch and open a pull request
* Once the PR has been merged immediately create a new tag `vX.Y.Z` and push it to GitHub
  * This will kick-off the goreleaser GitHub Action to build and push the release
  * If GHA is unavailable run `goreleaser release --release-notes <(go run helpers/changelog/main.go)` (with a valid `GITHUB_TOKEN` in your env) locally
* Check the [release](https://github.com/gopasspw/gopass/releases) on GitHub and verify the release notes

Some of these steps have been automated so it boils down to:

* Determine the next version, a patch increase is assumed. Otherwise provide the new version to the script.

```
$ go run helpers/release/main.go [X.Y.Z]
$ git push <your-fork> release/vX.Y.Z
# Open a PR, once it's merged
$ git co master
$ git pull origin master
$ git tag -s vX.Y.Z
$ git push origin vX.Y.Z
```

### Reproducible Builds

`gopass` supports [reproducible builds](https://reproducible-builds.org/). When
building from git [`SOURCE_DATE_EPOCH`](https://reproducible-builds.org/docs/source-date-epoch/)
can be used to override the compile date, .e.g `SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct)`.
When building a release `goreleaser` will automatically use the exact timestamp
of the last commit.

Internal paths are stripped using `-trimpath` and appropriate `-ldflags` (e.g. 
`-s`, `-w`). See the Makefile header for the exact set of flags.

