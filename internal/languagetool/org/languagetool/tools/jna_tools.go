package tools

import "os"

// SetJnaBugWorkaroundProperty ports JnaTools.setBugWorkaroundProperty.
// Sets jna.nosys=true via environment (process-local); call only from main-like entry points.
func SetJnaBugWorkaroundProperty() {
	_ = os.Setenv("jna.nosys", "true")
}
