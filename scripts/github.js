function search(query) {
	var url = "https://api.github.com/search/repositories?q=" + query + "&sort=stars&order=desc"
	var json = JSON.parse(httpGet(url));
	var result = [];
	json.items.forEach(function(elm) {
		log("search: " + elm.full_name.replace("/", "_"));
		var url = "https://api.github.com/repos/" + elm.full_name + "/contents"
		result.push({type: "dir", name: elm.stargazers_count + "_" + elm.language + "_" + elm.full_name.replace("/", "_"), url: url, size: "4096"});
	});
	return result;
}

function list(url) {
	var data = httpGet(url);
	var json = JSON.parse(data);
	var result = [];
	if (json.tree == null) {
		json.forEach(function(elm) {
			if (elm.type == "file") {
				result.push({type: "file", name: elm.name, url: elm.git_url, size: elm.size.toString()});
			} else if (elm.type == "dir") {
				result.push({type: "dir", name: elm.name, url: elm.git_url, size: elm.size.toString()});
			}
		});
	} else {
		json.tree.forEach(function(elm) {
			if (elm.type == "blob") {
				result.push({type: "file", name: elm.path, url: elm.url, size: elm.size.toString()});
			} else if (elm.type == "tree") {
				result.push({type: "dir", name: elm.path, url: elm.url, size:"4096"});
			}
		});
	}
	return result;
}

function download(url) {
	var json = JSON.parse(httpGet(url));
	return atob(json.content);
}
