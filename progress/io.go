/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package progress

import "io"

type IOTracker interface {
	SetSize(int64)
	Proxy(io.Reader) io.Reader
}
