package assign

import "regexp"

const pluginName = "assign"

var (
	assignRegexp = regexp.MustCompile(`(?mi)^/(un)?assign(( @?[-\w]+?)*)\s*$`)
	CCRegexp     = regexp.MustCompile(`(?mi)^/(un)?cc(( +@?[-/\w]+?)*)\s*$`)
)
