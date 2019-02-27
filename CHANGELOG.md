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
