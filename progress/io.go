/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package progress

import "io"

type IOTracker interface {
	Reset(string)
	SetSize(int64)
	Writer() io.Writer
}
