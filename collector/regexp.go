package collector

import "regexp"

var (
	// Regexp to parse lmutil output.
	lmutilVersionRegex = regexp.MustCompile(
		`^lmstat (?P<version>v[\d\.]+) build (?P<build>\d+) (?P<arch>[\w]+)`)
	lmutilLicenseServersRegex = regexp.MustCompile(
		`^License server status: (?P<servers>[\w,.@\-]+)`)
	lmutilLicenseServerStatusRegex = regexp.MustCompile(
		`(?P<fqdn>[\w\.\-]+): license server (?P<status>\w+)(?P<master>\s` +
			`\(MASTER\))? (?P<version>v[\d\.]+)$`)
	lmutilLicenseVendorStatusRegex = regexp.MustCompile(
		`^\s+(?P<vendor>\w+): (?P<status>UP|DOWN) (?P<version>v[\d\.]+)$`)
	lmutilLicenseFeatureUsageRegex = regexp.MustCompile(
		`^Users of (?P<name>.*):\s+\(Total of (?P<issued>\d+) \w+ issued\;\s+` +
			`Total of (?P<used>\d+) \w+ in use\)$`)
	lmutilLicenseFeatureUsageUserRegex = regexp.MustCompile(
		`^\s+(?P<user>[\w[:print:]]+) [\w\-\.]+ [[:print:]]+ ?\(v[\w\.]+\) \([\w\-\.]+\/\d+ ` +
			`\d+\)\, start \w+ \d+\/\d+ \d+\:\d+(\,\s(?P<licenses>\d+)\s\w+|)` +
			`(\s+\(linger\:\s\d+\s\/\s\d+\))?$`)
	lmutilLicenseFeatureUsageUser2Regex = regexp.MustCompile(
		`^\s+(?P<user>[\w[:print:]]+) [\w\-\.]+ ?\(v[\w\.]+\) \([\w\-\.]+\/\d+ ` +
			`\d+\)\, start \w+ \d+\/\d+ \d+\:\d+(\,\s(?P<licenses>\d+)\s\w+|)` +
			`(\s+\(linger\:\s\d+\s\/\s\d+\))?$`)
	lmutilLicenseFeatureGroupReservRegex = regexp.MustCompile(
		`^(\s+|)(?P<reservation>\d+)\s+\w+\s+for\s+(HOST_GROUP|GROUP)\s+` +
			`(?P<group>\w+).*$`)
	// lmutil lmstat -c port@hostname -i.
	lmutilLicenseFeatureExpRegex = regexp.MustCompile(
		`^(?P<feature>[[:graph:]]+)\s+(?P<version>[\d\.]+)\s+` +
			`(?P<licenses>\d+)\s+(?P<expires>[\w\-]+)\s+(?P<vendor>\w+)$`)
	// lmutil lmstat -c port@hostname -i case with columns expired and vendors
	// switched #28.
	lmutilLicenseFeatureExpRegex2 = regexp.MustCompile(
		`^(?P<feature>[[:graph:]]+)\s+(?P<version>[\d\.]+)\s+` +
			`(?P<licenses>\d+)\s+(?P<vendor>\w+)\s+(?P<expires>[\w\-\s\(\)]+)$`)
)
