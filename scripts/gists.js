function root() {
	return [{type: "dir", name: "my", url: "https://api.github.com/users/wiegelmann/gists"}];
}

function list(url) {
	var json = JSON.parse(httpGet(url));
	var result = [];
	json.forEach(function(elm) {
		var key = Object.keys(elm.files)[0];
		var child = elm.files[key];
		result.push({type: "file", name: child.filename, url: child.raw_url, size: child.size.toString()});
	});
	return result;
}

function download(url) {
	return httpGet(url);
}
