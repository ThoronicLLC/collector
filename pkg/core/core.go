package core

// MaxLogSize specifies the max size for a log (in this case 5MB)
// we do this because some of our output plugins have a limit on per
// entry size.
//
// Max size per log = 5MB
var MaxLogSize = 5 * 1024 * 1024
