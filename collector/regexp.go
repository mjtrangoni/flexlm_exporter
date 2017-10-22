package collector

import "regexp"

var (
	// Regexp to parse lmutil output.
	lmutilVersionRegex = regexp.MustCompile(
		`^lmstat (?P<version>v[\d\.]+) build (?P<build>\d+) (?P<arch>[\w\_]+)`)
	lmutilLicenseServersRegex = regexp.MustCompile(
		`^License server status: (?P<servers>[\w\,\.\@]+)`)
	lmutilLicenseServerStatusRegex = regexp.MustCompile(
		`^(?P<fqdn>[\w\.]+): license server (?P<status>\w+)(?P<master>\s` +
			`\(MASTER\))? (?P<version>v[\d\.]+)$`)
	lmutilLicenseVendorStatusRegex = regexp.MustCompile(
		`^\s+(?P<vendor>\w+): (?P<status>UP|DOWN) (?P<version>v[\d\.]+)$`)
	lmutilLicenseFeatureUsageRegex = regexp.MustCompile(
		`^Users of (?P<name>.*):\s+\(Total of (?P<issued>\d+) licenses issued` +
			`\;\s+Total of (?P<used>\d+) licenses in use\)$`)
	lmutilLicenseFeatureUsageUserRegex = regexp.MustCompile(
		`^\s+(?P<user>\w+) [\w\-]+ [\w\/\-]+ \(v[\w\.]+\) \([\w\-\.]+\/\d+ ` +
			`\d+\)\, start \w+ \d+\/\d+ \d+\:\d+(\,\s(?P<licenses>\d+) \w+|)$`)
	//      8 RESERVATIONs for GROUP GROUP8 (host3.domain.net/27002)
	//        8 RESERVATIONs for GROUP PWTDPP (vumca538.rd.corpintra.net/27002)
	//8 RESERVATIONs for GROUP NCSCSD (vumca538.rd.corpintra.net/27002)24
	lmutilLicenseFeatureGroupReservRegex = regexp.MustCompile(
		`^(\s+|)(?P<reservation>\d+)\s+\w+\s+for\s+GROUP\s+(?P<group>\w+).*$`)
)
