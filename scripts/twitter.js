function root() {
	var result = [];
	result.push({type: "dir", name: "timeline", url: "https://api.twitter.com/1.1/statuses/home_timeline.json?count=50"});
	return result;
}

function list(url) {
	var key = "";
	var keysecret = "";
	var token = "";
	var tokensecret = "";
	var data = httpGetOAuth1(url, key, keysecret, token, tokensecret);
	var json = JSON.parse(data);
	var result = [];
	var content = "";
	json.forEach(function(elm) {
		content += elm.user.name + ": " + elm.text + "\n---\n";
	});
	result.push({type: "file", name: "home.txt", data: content, size: content.length.toString()});
	return result;
}

function download(url) {
	var json = JSON.parse(httpGet(url));
	return atob(json.content);
}
