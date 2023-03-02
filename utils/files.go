package utils

func IsCodeFile(typeFile string) bool {
	if typeFile == ".py" || typeFile == ".js" || typeFile == ".html" || typeFile == "css" || typeFile == ".c" {
		return true
	}
	return false
}

func GetCodeFileMime(typeFile string) string {
	switch typeFile {
	case ".py":
		return "text/x-python"
	case ".js":
		return "application/javascript"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".c":
		return "text/x-c"
	default:
		return ""
	}
}
