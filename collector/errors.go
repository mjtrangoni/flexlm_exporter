package collector

// The original error codes are converted to unsigned integers,
// e.g. -15 = 241 (-15 + 256)
// Reference: http://www.opendtect.org/lic/doc/endusermanual/chap13.htm
var errorDescriptionString = map[string]string{
	"exit status 255": "Cannot find license file.",
	"exit status 254": "Invalid license file syntax.",
	"exit status 253": "No server for this feature.",
	"exit status 252": "Licensed number of users already reached.",
	"exit status 251": "No such feature exists.",
	"exit status 250": "No TCP/IP port number in license file and FLEXlm service does not exist. (pre-v6 only)",
	"exit status 249": "No socket connection to license manager service.",
	"exit status 248": "Invalid (inconsistent) license key or signature. " +
		"The license key/signature and data for the feature do not match. " +
		"This usually happens when a license file has been altered.",
	"exit status 247": "Invalid host. The hostid of this system does not " +
		"match the hostid specified in the license file.",
	"exit status 246": "Feature has expired.",
	"exit status 245": "Invalid date format in license file.",
	"exit status 244": "Invalid returned data from license server.",
	"exit status 243": "No SERVER lines in license file.",
	"exit status 242": "Cannot find SERVER host name in network database. " +
		"The lookup for the host name on the SERVER line in the license " +
		"file failed. This often happens when NIS or DNS or the hosts " +
		"file is incorrect. Workaround: Use IP address " +
		"(e.g., 123.456.789.123) instead of host name.",
	"exit status 241": "Cannot connect to license server. The server (lmgrd) " +
		"has not been started yet, or the wrong port@host or license file is" +
		" being used, or the TCP/IP port or host name in the license file has been changed.",
	"exit status 240": "Cannot read data from license server.",
	"exit status 239": "Cannot write data to license server.",
	"exit status 238": "License server does not support this feature.",
	"exit status 237": "Error in select system call.",
	"exit status 235": "License file does not support this version.",
	"exit status 234": "Feature checkin failure detected at license server.",
	"exit status 233": "License server temporarily busy (new server connecting).",
	"exit status 232": "Users are queued for this feature.",
	"exit status 231": "License server does not support this version of this feature.",
	"exit status 230": "Request for more licenses than this feature supports.",
	"exit status 227": "Cannot find ethernet device.",
	"exit status 226": "Cannot read license file.",
	"exit status 225": "Feature start date is in the future.",
	"exit status 224": "No such attribute.",
	"exit status 223": "Bad encryption handshake with daemon.",
	"exit status 222": "Clock difference too large between client and server.",
	"exit status 221": "In the queue for this feature.",
	"exit status 220": "Feature database corrupted in daemon.",
	"exit status 219": "Duplicate selection mismatch for this feature. Obsolete with v8.0+ vendor daemon.",
	"exit status 218": "User/host on EXCLUDE list for feature.",
	"exit status 217": "User/host not on INCLUDE list for feature.",
	"exit status 216": "Cannot locate dynamic memory.",
	"exit status 215": "Feature was never checked out.",
	"exit status 214": "Invalid parameter.",
	"exit status 209": "Clock setting check not available in daemon.",
	"exit status 204": "FLEXlm vendor daemon did not respond within timeout interval.",
	"exit status 203": "Checkout request rejected by vendor-defined checkout filter.",
	"exit status 202": "No FEATURESET line in license file.",
	"exit status 201": "Incorrect FEATURESET line in license file.",
	"exit status 200": "Cannot compute FEATURESET data from license file.",
	"exit status 199": "socket() call failed.",
	"exit status 197": "Message checksum failure.",
	"exit status 196": "Server message checksum failure.",
	"exit status 195": "Cannot read license file data from server.",
	"exit status 194": "Network software (TCP/IP) not available.",
	"exit status 193": "You are not a license administrator.",
	"exit status 192": "lmremove request before the minimum lmremove interval.",
	"exit status 189": "No licenses to borrow.",
	"exit status 188": "License BORROW support not enabled.",
	"exit status 187": "FLOAT_OK can’t run standalone on SERVER.",
	"exit status 185": "Invalid TZ environment variable.",
	"exit status 183": "Local checkout filter rejected request.",
	"exit status 182": "Attempt to read beyond end of license file path.",
	"exit status 181": "SYS$SETIMR call failed (VMS).",
	"exit status 180": "Internal FLEXlm error—please report to Macrovision.",
	"exit status 179": "Bad version number must be floating-point number with no letters.",
	"exit status 174": "Invalid PACKAGE line in license file.",
	"exit status 173": "FLEXlm version of client newer than server.",
	"exit status 172": "USER_BASED license has no specified users - see server log.",
	"exit status 171": "License server doesn’t support this request.",
	"exit status 169": "Checkout exceeds MAX specified in options file.",
	"exit status 168": "System clock has been set back.",
	"exit status 167": "This platform not authorized by license.",
	"exit status 166": "Future license file format or misspelling in license file. " +
		"The file was issued for a later version of FLEXlm than this program understands.",
	"exit status 165": "ENCRYPTION_SEEDS are non-unique.",
	"exit status 164": "Feature removed during lmreread, or wrong SERVER line hostid.",
	"exit status 163": "This feature is available in a different license pool. This is a " +
		"warning condition. The server has pooled one or more INCREMENT lines into a " +
		"single pool, and the request was made on an INCREMENT line that has been pooled.",
	"exit status 162": "Attempt to generate license with incompatible attributes.",
	"exit status 161": "Network connect to this_host failed. Change this_host on the SERVER " +
		"line in the license file to the actual host name.",
	"exit status 160": "Server machine is down or not responding. See the system administrator " +
		"about starting the server, or make sure that you’re referring to the right host " +
		"(see LM_LICENSE_FILE environment variable).",
	"exit status 159": "The desired vendor daemon is down. 1) Check the lmgrd log file, or 2) Try lmreread.",
	"exit status 158": "This FEATURE line can’t be converted to decimal format.",
	"exit status 157": "The decimal format license is typed incorrectly.",
	"exit status 156": "Cannot remove a linger license.",
	"exit status 155": "All licenses are reserved for others. The system administrator " +
		"has reserved all the licenses for others. Reservations are made in the " +
		"options file. The server must be restarted for options file changes to take effect.",
	"exit status 154": "A FLEXid borrow error occurred.",
	"exit status 153": "Terminal Server remote client not allowed.",
	"exit status 152": "Cannot borrow that long.",
	"exit status 150": "License server out of network connections. The vendor daemon " +
		"can't handle any more users. See the debug log for further information.",
	"exit status 146": "Dongle not attached, or can’t read dongle. Either the hardware dongle " +
		"is unattached, or the necessary software driver for this dongle type is not installed.",
	"exit status 144": "Missing dongle driver. In order to read the dongle hostid, the " +
		"correct driver must be installed. These drivers are available at www.macrovision.com " +
		"or from your software vendor.",
	"exit status 143": "Two FLEXlock checkouts attempted. Only one checkout is allowed with " +
		"FLEXlock-enabled applications.",
	"exit status 142": "SIGN= keyword required, but missing from license. This is probably " +
		"because the license is older than the application. You need to obtain a SIGN= " +
		"version of this license from your vendor.",
	"exit status 141": "Error in Public Key package.",
	"exit status 140": "CRO not supported for this platform.",
	"exit status 139": "BORROW failed.",
	"exit status 138": "BORROW period has expired.",
	"exit status 137": "lmdown and lmreread must be run on license server machine.",
	"exit status 136": "Cannot lmdown the server when licenses are borrowed.",
	"exit status 135": "FLOAT_OK license must have exactly one dongle hostid.",
	"exit status 134": "Unable to delete local borrow info.",
	"exit status 133": "Support for returning a borrowed license early is not enabled. The vendor " +
		"must have enabled support for this feature in the vendor daemon. Contact the vendor for further details.",
	"exit status 132": "An error occurred while returning a borrowed license to the server.",
	"exit status 131": "Attempt to checkout just a PACKAGE. Need to also checkout a feature.",
	"exit status 130": "Error initializing a composite hostid.",
	"exit status 129": "A hostid needed for the composite hostid is missing or invalid.",
	"exit status 128": "Error, borrowed license doesn't match any known server license.",
}
