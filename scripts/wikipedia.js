function search(query) {
	var url = "https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=" + query + "&srlimit=50&format=json"
	var json = JSON.parse(httpGet(url));
	var result = [];
	json.query.search.forEach(function(elm) {
		var file = "https://en.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exlimit=1&explaintext&pageids=" + elm.pageid + "&&formatversion=2&redirects="
		result.push({type: "file", name: elm.title + ".txt", url: file, size: "4096"});
	});
	return result;
}

function download(url) {
	var json = JSON.parse(httpGet(url));
	return json.query.pages[0].extract;
}
