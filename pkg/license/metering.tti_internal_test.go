// Copyright © 2019 The Things Industries B.V.

package license

var NewMeteringSetup = newMeteringSetup

func (s *meteringSetup) ReplaceReporter(r MeteringReporter) {
	s.reporter = r
}
