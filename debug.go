package goapi

import (
	"strings"
)

func debugPrintRouter(log Logger, paths []*pathInfo) {
	if log == nil {
		return
	}
	log.Debug("All routes:")
	headMethod := "METHODS"
	headPath := "PATH"
	headPos := "POSITION"
	maxMethodLen := len(headMethod)
	maxPathLen := len(headPath)
	maxPosLen := len(headPos)
	for _, path := range paths {
		methodLen := len(strings.Join(path.methods, ","))
		if methodLen > maxMethodLen {
			maxMethodLen = methodLen
		}
		pathLen := len(strings.Join(path.paths, ","))
		if path.isSwagger {
			pathLen = len(path.paths[0])
		}
		if pathLen > maxPathLen {
			maxPathLen = pathLen
		}
		posLen := len(path.pos)
		if posLen > maxPosLen {
			maxPosLen = posLen
		}
	}
	log.Debug(strings.Repeat("-", maxMethodLen+maxPathLen+maxPosLen+10))
	log.Debug("| %v | %v | %v |", spanFill(headMethod, len(headMethod), maxMethodLen),
		spanFill(headPath, len(headPath), maxPathLen), spanFill(headPos, len(headPos), maxPosLen))
	for _, path := range paths {
		method := strings.Join(path.methods, ",")
		p := strings.Join(path.paths, ",")
		if path.isSwagger {
			p = path.paths[0]
		}
		log.Debug("| %v | %v | %v |", spanFill(method, len(method), maxMethodLen),
			spanFill(p, len(p), maxPathLen), spanFill(path.pos, len(path.pos), maxPosLen))
	}
	log.Debug(strings.Repeat("-", maxMethodLen+maxPathLen+maxPosLen+10))
}
