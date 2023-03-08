package util

func JoinUrl(base string, path string) string {
	url := base
	if url[len(url)-1:] == "/" {
		if path[:1] == "/" {
			url = url + path[1:]
		} else {
			url = url + path
		}
	} else {
		if path[:1] == "/" {
			url = url + path
		} else {
			url = url + "/" + path
		}
	}

	return url
}
