## v0.0.8 / unreleased

 * [ENHANCEMENT] Compile for darwin-arm64 and darwin-amd64.
 * [ENHANCEMENT] Consolidate windows and linux support #53. Thanks @SuperSonicFlea.
 * [ENHANCEMENT] Replace EOL CentOS8 docker image with RockyLinux8.

## v0.0.7 / 2021-09-25

 * [ENHANCEMENT] Replace deprecated `prometheus/common/log` with `go-kit/log`.
 * [BUGFIX] Count licenses separately for different versions, which fix
   "flexlm_feature_used_users" implementation. Thanks @TobiasKadelka.

## v0.0.6 / 2021-07-14

 * [ENHANCEMENT] Add version flag to "flexlm_feature_used_users". Thanks @TobiasKadelka.
 * [ENHANCEMENT] CI: Move from TravisCI to GitHub Actions.
 * [BUGFIX] Fix #33 "Error during parsing expiration date". Thanks @99rgosse.

## v0.0.5 / 2020-10-15

 * [ENHANCEMENT] Handle case for switched columns on expirations output.
 * [ENHANCEMENT] Build Docker image on CentOS8.
 * [ENHANCEMENT] Use goroutines for each license and add features aggregate
   expiration. Thanks @treydock.
 * [ENHANCEMENT] Switch to go modules and yaml.v3. Thanks @knweiss.
 * [ENHANCEMENT] First crossbuild support.

## v0.0.4 / 2019-02-27

 * [ENHANCEMENT] Changed Regex to include FQDNs.
 * [ENHANCEMENT] Handle the case where no terminal devices are present in
   `lmstat` output.

## v0.0.3 / 2018-10-10

* [ENHANCEMENT] Handle repeated feature expirations better.
* [BUGFIX] rename `flexlm_lmstat_feature_expiration_seconds` to
  `flexlm_feature_expiration_seconds`

## v0.0.2 / 2018-10-01

* [ENHANCEMENT] Add encrypted displays output support.
* [ENHANCEMENT] Force `lmutil` to run with `LANG=C`.
* [ENHANCEMENT] Update vendoring.
* [ENHANCEMENT] Improve testing.
* [ENHANCEMENT] expose `-i` expiration dates in seconds with new
  `lmstat_feature_exp` collector.
* [BUGFIX] Fix `feature_to_include` logic error.

## v0.0.1 / 2018-06-12

 * First working version.
