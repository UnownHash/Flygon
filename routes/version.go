package routes

import (
	"os"
	"runtime/debug"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}()

var Version = func() string {
	if info, err := os.Open("VERSION"); err == nil {
		bytes := make([]byte, 5)
		_, err = info.Read(bytes)
		if err != nil {
			return "undefined"
		}
		return string(bytes)
	}
	return "undefined"
}()
