package goapi

import (
	"fmt"
	"strconv"
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
	methodMap := map[string]int{}
	countRouter := 0
	var methods []string
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
		posLen := len(debugPos(path))
		if posLen > maxPosLen {
			maxPosLen = posLen
		}
		for _, method := range path.methods {
			if methodMap[method] == 0 {
				methods = append(methods, method)
			}
			methodMap[method] += len(path.paths)
			countRouter += len(path.paths)
		}
	}
	if countRouter > 0 {
		lineLen := maxMethodLen + maxPathLen + maxPosLen + 10
		log.Debug(strings.Repeat("-", lineLen))
		log.Debug("| %v | %v | %v |", spanFill(headMethod, len(headMethod), maxMethodLen),
			spanFill(headPath, len(headPath), maxPathLen), spanFill(headPos, len(headPos), maxPosLen))
		for _, path := range paths {
			method := strings.Join(path.methods, ",")
			p := strings.Join(path.paths, ",")
			if path.isSwagger {
				p = path.paths[0]
			}
			pos := debugPos(path)
			log.Debug("| %v | %v | %v |", spanFill(method, len(method), maxMethodLen),
				spanFill(p, len(p), maxPathLen), spanFill(pos, len(pos), maxPosLen))
		}
		log.Debug(strings.Repeat("-", lineLen))
		var countMethods []string
		for _, method := range methods {
			newMethod := method + "(" + strconv.Itoa(methodMap[method]) + ")"
			countMethods = append(countMethods, newMethod)
		}
		debugPrintMethod(log, countMethods, countRouter, lineLen)
	} else {
		log.Debug("No routing available")
	}
}

func debugPrintMethod(log Logger, methods []string, countRouter, totalLen int) {
	leftLen := 15
	rightLen := totalLen - leftLen - 2
	val := ""
	methodCount := 0
	for index, method := range methods {
		if index < len(methods)-1 {
			method += ", "
		}
		if len(val+method) <= rightLen {
			val += method
			continue
		}
		if methodCount == 0 {
			log.Debug("| MethodCount: " + spanFill(val, len(val), rightLen) + " |")
		} else {
			log.Debug(spanFill("| ", 2, leftLen) + spanFill(val, len(val), rightLen) + " |")
		}
		val = method
		methodCount++
	}
	log.Debug(spanFill("| ", 2, leftLen) + spanFill(val, len(val), rightLen) + " |")
	routerCount := strconv.Itoa(countRouter)
	log.Debug("| RouterCount: " + spanFill(routerCount, len(routerCount), rightLen) + " |")
	log.Debug(strings.Repeat("-", totalLen))
}

func debugPos(path *pathInfo) string {
	pos := path.pos
	if path.isSwagger {
		pos += " (docs)"
	}
	if path.inFs != nil {
		pos += " (fs)"
	}
	if len(path.middlewares) > 0 {
		pos += fmt.Sprintf(" (%v Middleware)", len(path.middlewares))
	}
	securityCount := 0
	for _, item := range path.inParams {
		if inArray(item.inType, []InType{
			inTypeSecurityHTTPBearer,
			inTypeSecurityHTTPBearerJWT,
			inTypeSecurityHTTPBasic,
			inTypeSecurityApiKey,
		}) {
			securityCount++
		}
	}
	if securityCount > 0 {
		pos += fmt.Sprintf(" (%v Security)", securityCount)
	}
	return pos
}
