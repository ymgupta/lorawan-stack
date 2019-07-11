// Copyright © 2019 The Things Industries B.V.

package ttnmage

func init() {
	if goTags == "" {
		goTags = "tti"
	}
	goBinaries = append(goBinaries, "tti-lw-cli", "tti-lw-stack")
}
